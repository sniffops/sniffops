package trace

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SanitizeOutput은 민감한 데이터를 마스킹하여 trace 저장소에 안전하게 저장할 수 있도록 합니다.
//
// 마스킹 대상:
// - Secret, ConfigMap data 필드
// - API key, password, token 패턴
// - 환경 변수의 민감한 값
// - URL의 credential
func SanitizeOutput(output string) string {
	// Empty output은 그대로 반환
	if output == "" {
		return output
	}

	// JSON 파싱 시도 (K8s resource는 주로 JSON 형식)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err == nil {
		// JSON인 경우 구조적으로 sanitize
		sanitized := sanitizeJSONMap(data)
		result, _ := json.Marshal(sanitized)
		return string(result)
	}

	// JSON이 아닌 경우 텍스트 기반 sanitize (예: logs, exec output)
	return sanitizeText(output)
}

// sanitizeJSONMap은 JSON map을 재귀적으로 순회하며 민감한 필드를 마스킹합니다.
func sanitizeJSONMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		lowerKey := strings.ToLower(key)

		// Secret이나 ConfigMap의 data 필드 전체 마스킹
		if (lowerKey == "data" || lowerKey == "stringdata") {
			// 상위 객체의 kind 확인 (완벽하진 않지만 일반적인 케이스 처리)
			if kind, ok := data["kind"].(string); ok {
				if kind == "Secret" || kind == "ConfigMap" {
					// data 필드의 모든 값을 [REDACTED]로 대체
					if dataMap, ok := value.(map[string]interface{}); ok {
						redactedData := make(map[string]interface{})
						for k := range dataMap {
							redactedData[k] = "[REDACTED]"
						}
						result[key] = redactedData
						continue
					}
				}
			}
		}

		// 민감한 키 이름 패턴 (password, token, key, secret 등)
		if isSensitiveKey(lowerKey) {
			result[key] = "[REDACTED]"
			continue
		}

		// 재귀적으로 처리
		switch v := value.(type) {
		case map[string]interface{}:
			result[key] = sanitizeJSONMap(v)
		case []interface{}:
			result[key] = sanitizeJSONArray(v)
		case string:
			// 문자열 값도 패턴 검사
			result[key] = sanitizeText(v)
		default:
			result[key] = value
		}
	}

	return result
}

// sanitizeJSONArray는 JSON array를 재귀적으로 순회하며 민감한 필드를 마스킹합니다.
func sanitizeJSONArray(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			// 특별 케이스: K8s env 항목 (name + value 구조)
			// env의 name이 민감한 키워드면 value도 redact
			if envName, hasName := v["name"].(string); hasName {
				if _, hasValue := v["value"]; hasValue {
					if isSensitiveKey(strings.ToLower(envName)) {
						sanitized := make(map[string]interface{})
						for k, val := range v {
							if k == "value" {
								sanitized[k] = "[REDACTED]"
							} else {
								sanitized[k] = val
							}
						}
						result[i] = sanitized
						continue
					}
				}
			}
			
			result[i] = sanitizeJSONMap(v)
		case []interface{}:
			result[i] = sanitizeJSONArray(v)
		case string:
			result[i] = sanitizeText(v)
		default:
			result[i] = value
		}
	}

	return result
}

// isSensitiveKey는 키 이름이 민감한 정보를 포함하는지 확인합니다.
func isSensitiveKey(key string) bool {
	sensitiveKeywords := []string{
		"password",
		"passwd",
		"pwd",
		"secret",
		"token",
		"apikey",
		"api_key",
		"api-key",
		"credential",
		"auth",
		"authorization",
		"bearer",
		"private",
		"privatekey",
		"private_key",
		"private-key",
		"cert",
		"certificate",
		"key",
	}

	for _, keyword := range sensitiveKeywords {
		if strings.Contains(key, keyword) {
			return true
		}
	}

	return false
}

// sensitivePatterns는 텍스트에서 민감한 패턴을 찾기 위한 정규표현식 목록입니다.
var sensitivePatterns = []*regexp.Regexp{
	// API keys (일반적인 형식)
	regexp.MustCompile(`(?i)(api[_-]?key\s*[:=]\s*)["']?([a-zA-Z0-9_\-]{20,})["']?`),
	
	// API keys (sk-, pk- 등의 prefix - standalone)
	regexp.MustCompile(`(?i)\b(sk|pk|api)[_-]?[a-zA-Z0-9]{15,}\b`),
	
	// Environment variable pattern (e.g., "API_KEY environment variable: value")
	regexp.MustCompile(`(?i)(password|api[_-]?key|token|secret)\s+[a-z\s]*:\s*([a-zA-Z0-9_\-]{3,})`),

	// Bearer tokens
	regexp.MustCompile(`(?i)(bearer\s+)([a-zA-Z0-9_\-\.]{20,})`),

	// AWS credentials
	regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key[_-]?id\s*[:=]\s*)["']?([A-Z0-9]{20})["']?`),
	regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*)["']?([a-zA-Z0-9/+=]{40})["']?`),

	// Passwords in various formats
	regexp.MustCompile(`(?i)(password\s*[:=]\s*)["']?([^\s"',}]{3,})["']?`),
	regexp.MustCompile(`(?i)(passwd\s*[:=]\s*)["']?([^\s"',}]{3,})["']?`),
	regexp.MustCompile(`(?i)(pwd\s*[:=]\s*)["']?([^\s"',}]{3,})["']?`),

	// Generic secrets
	regexp.MustCompile(`(?i)(secret\s*[:=]\s*)["']?([^\s"',}]{3,})["']?`),

	// Tokens
	regexp.MustCompile(`(?i)(token\s*[:=]\s*)["']?([a-zA-Z0-9_\-\.]{20,})["']?`),

	// URLs with credentials (http://user:pass@host, postgres://user:pass@host, etc.)
	regexp.MustCompile(`(?i)((?:https?|postgres|postgresql|mysql|mongodb)://)([^:]+):([^@]+)@`),

	// Private keys (entire block)
	regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),

	// Base64-encoded secrets (heuristic: long base64 strings in value position)
	regexp.MustCompile(`(?i)(secret|password|token|key)\s*[:=]\s*["']?([A-Za-z0-9+/]{40,}={0,2})["']?`),
}

// sanitizeText는 텍스트에서 민감한 패턴을 마스킹합니다.
func sanitizeText(text string) string {
	result := text

	for _, pattern := range sensitivePatterns {
		patternStr := pattern.String()
		
		// Private key의 경우 전체 블록 제거
		if strings.Contains(patternStr, "PRIVATE KEY") {
			result = pattern.ReplaceAllString(result, "[PRIVATE KEY REDACTED]")
			continue
		}

		// URL credential의 경우 (http, https, postgres, mysql, mongodb 등)
		if strings.Contains(patternStr, "postgres") || strings.Contains(patternStr, "https?") {
			result = pattern.ReplaceAllString(result, "$1***:***@")
			continue
		}

		// sk-, pk- 등의 API key prefix (standalone pattern)
		if strings.Contains(patternStr, `\b(sk|pk|api)`) {
			result = pattern.ReplaceAllString(result, "[REDACTED]")
			continue
		}

		// 일반적인 key=value 패턴
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// 첫 번째 그룹(키 부분)은 유지하고 값만 마스킹
			if strings.Contains(match, ":") || strings.Contains(match, "=") {
				parts := regexp.MustCompile(`[:=]`).Split(match, 2)
				if len(parts) == 2 {
					return parts[0] + "=[REDACTED]"
				}
			}
			return "[REDACTED]"
		})
	}

	return result
}
