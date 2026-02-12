package trace

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeOutput_EmptyString(t *testing.T) {
	result := SanitizeOutput("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestSanitizeOutput_SecretData(t *testing.T) {
	// Kubernetes Secret with base64-encoded data
	input := `{
		"apiVersion": "v1",
		"kind": "Secret",
		"metadata": {
			"name": "db-password",
			"namespace": "production"
		},
		"data": {
			"password": "cGFzc3dvcmQxMjM=",
			"username": "YWRtaW4="
		}
	}`

	result := SanitizeOutput(input)

	// Verify data fields are redacted
	if strings.Contains(result, "cGFzc3dvcmQxMjM=") {
		t.Error("Password data should be redacted")
	}
	if strings.Contains(result, "YWRtaW4=") {
		t.Error("Username data should be redacted")
	}

	// Verify structure is preserved
	var sanitized map[string]interface{}
	if err := json.Unmarshal([]byte(result), &sanitized); err != nil {
		t.Fatalf("Failed to parse sanitized output: %v", err)
	}

	// Check data field is redacted
	data := sanitized["data"].(map[string]interface{})
	if data["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be [REDACTED], got %v", data["password"])
	}
	if data["username"] != "[REDACTED]" {
		t.Errorf("Expected username to be [REDACTED], got %v", data["username"])
	}

	// Metadata should be intact
	metadata := sanitized["metadata"].(map[string]interface{})
	if metadata["name"] != "db-password" {
		t.Errorf("Metadata name should not be redacted, got %v", metadata["name"])
	}
}

func TestSanitizeOutput_ConfigMapData(t *testing.T) {
	input := `{
		"apiVersion": "v1",
		"kind": "ConfigMap",
		"metadata": {
			"name": "app-config"
		},
		"data": {
			"api_key": "sk-1234567890abcdef",
			"database_url": "postgres://user:pass@localhost:5432/db"
		}
	}`

	result := SanitizeOutput(input)

	// ConfigMap data should also be redacted
	if strings.Contains(result, "sk-1234567890abcdef") {
		t.Error("API key in ConfigMap should be redacted")
	}
	if strings.Contains(result, "postgres://user:pass@") {
		t.Error("Database URL with credentials should be redacted")
	}
}

func TestSanitizeOutput_SensitiveKeys(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		contains []string // strings that should NOT be in output
	}{
		{
			name: "password field",
			input: `{
				"config": {
					"password": "secret123"
				}
			}`,
			contains: []string{"secret123"},
		},
		{
			name: "token field",
			input: `{
				"auth": {
					"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
				}
			}`,
			contains: []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		},
		{
			name: "api_key field",
			input: `{
				"settings": {
					"api_key": "AKIAIOSFODNN7EXAMPLE"
				}
			}`,
			contains: []string{"AKIAIOSFODNN7EXAMPLE"},
		},
		{
			name: "apiKey camelCase",
			input: `{
				"config": {
					"apiKey": "1234567890abcdef"
				}
			}`,
			contains: []string{"1234567890abcdef"},
		},
		{
			name: "private_key field",
			input: `{
				"tls": {
					"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIE..."
				}
			}`,
			contains: []string{"BEGIN RSA PRIVATE KEY"},
		},
		{
			name: "authorization field",
			input: `{
				"headers": {
					"authorization": "Bearer abc123xyz"
				}
			}`,
			contains: []string{"Bearer abc123xyz"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeOutput(tc.input)

			for _, secret := range tc.contains {
				if strings.Contains(result, secret) {
					t.Errorf("Secret value %q should be redacted in output: %s", secret, result)
				}
			}

			// Verify it's still valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("Sanitized output should be valid JSON: %v", err)
			}
		})
	}
}

func TestSanitizeText_APIKeys(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		contains string // string that should NOT be in output
	}{
		{
			name:     "API key pattern",
			input:    "api_key: sk-1234567890abcdef1234567890",
			contains: "sk-1234567890abcdef1234567890",
		},
		{
			name:     "API key with equals",
			input:    "API_KEY=AIzaSyD1234567890abcdef",
			contains: "AIzaSyD1234567890abcdef",
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			contains: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:     "AWS access key",
			input:    "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			contains: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:     "AWS secret key",
			input:    "AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			contains: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:     "Password field",
			input:    "password: mySecretP@ss123",
			contains: "mySecretP@ss123",
		},
		{
			name:     "Token field",
			input:    "token=ghp_1234567890abcdefghijklmnopqrstuvwxyz",
			contains: "ghp_1234567890abcdefghijklmnopqrstuvwxyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeText(tc.input)

			if strings.Contains(result, tc.contains) {
				t.Errorf("Secret value %q should be redacted in output: %s", tc.contains, result)
			}

			// Should contain [REDACTED] or similar
			if !strings.Contains(result, "[REDACTED]") && !strings.Contains(result, "***") {
				t.Errorf("Output should contain redaction marker: %s", result)
			}
		})
	}
}

func TestSanitizeText_URLCredentials(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		contains string // credential part that should NOT be in output
	}{
		{
			name:     "HTTP with credentials",
			input:    "database url: http://admin:password123@localhost:5432",
			contains: "admin:password123",
		},
		{
			name:     "HTTPS with credentials",
			input:    "connection: https://user:p@ssw0rd@api.example.com/v1",
			contains: "user:p@ssw0rd",
		},
		{
			name:     "PostgreSQL URL",
			input:    "DATABASE_URL=postgres://dbuser:dbpass@db.example.com:5432/mydb",
			contains: "dbuser:dbpass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeText(tc.input)

			if strings.Contains(result, tc.contains) {
				t.Errorf("URL credentials %q should be redacted in output: %s", tc.contains, result)
			}

			// Should contain masked credentials
			if !strings.Contains(result, "***:***@") {
				t.Errorf("URL should have masked credentials: %s", result)
			}
		})
	}
}

func TestSanitizeText_PrivateKey(t *testing.T) {
	input := `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef...
-----END RSA PRIVATE KEY-----
`

	result := sanitizeText(input)

	if strings.Contains(result, "BEGIN RSA PRIVATE KEY") {
		t.Error("Private key should be completely redacted")
	}
	if strings.Contains(result, "MIIEpAIBAAKCAQEA") {
		t.Error("Private key content should be redacted")
	}

	// Should contain redaction marker
	if !strings.Contains(result, "[PRIVATE KEY REDACTED]") {
		t.Errorf("Private key should have redaction marker: %s", result)
	}
}

func TestSanitizeOutput_NestedJSON(t *testing.T) {
	input := `{
		"metadata": {
			"name": "app-deployment"
		},
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"name": "app",
							"env": [
								{
									"name": "DATABASE_PASSWORD",
									"value": "super-secret-password"
								},
								{
									"name": "API_TOKEN",
									"value": "token-1234567890abcdef"
								},
								{
									"name": "LOG_LEVEL",
									"value": "info"
								}
							]
						}
					]
				}
			}
		}
	}`

	result := SanitizeOutput(input)

	// Secrets should be redacted
	if strings.Contains(result, "super-secret-password") {
		t.Error("DATABASE_PASSWORD value should be redacted")
	}
	if strings.Contains(result, "token-1234567890abcdef") {
		t.Error("API_TOKEN value should be redacted")
	}

	// Non-sensitive values should be preserved
	if !strings.Contains(result, "info") {
		t.Error("LOG_LEVEL value should not be redacted")
	}
	if !strings.Contains(result, "app-deployment") {
		t.Error("Deployment name should not be redacted")
	}
}

func TestSanitizeOutput_PlainText(t *testing.T) {
	// Non-JSON output (e.g., pod logs)
	input := `
Starting application...
Connecting to database with password=secretPass123
API_KEY environment variable: sk-abcdef1234567890
Server running on port 8080
`

	result := SanitizeOutput(input)

	if strings.Contains(result, "secretPass123") {
		t.Error("Password in logs should be redacted")
	}
	if strings.Contains(result, "sk-abcdef1234567890") {
		t.Error("API key in logs should be redacted")
	}

	// Non-sensitive content should be preserved
	if !strings.Contains(result, "Starting application") {
		t.Error("Non-sensitive log content should be preserved")
	}
	if !strings.Contains(result, "Server running on port 8080") {
		t.Error("Non-sensitive log content should be preserved")
	}
}

func TestSanitizeOutput_Base64Secrets(t *testing.T) {
	input := `{
		"credentials": {
			"secret": "dGhpc19pc19hX3NlY3JldF90b2tlbl90aGF0X2lzX3ZlcnlfbG9uZ19hbmRfYmFzZTY0X2VuY29kZWQ="
		}
	}`

	result := SanitizeOutput(input)

	// Long base64 string in secret field should be redacted
	if strings.Contains(result, "dGhpc19pc19hX3NlY3JldF90b2tlbl90aGF0X2lzX3ZlcnlfbG9uZ19hbmRfYmFzZTY0X2VuY29kZWQ=") {
		t.Error("Base64-encoded secret should be redacted")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	testCases := []struct {
		key       string
		sensitive bool
	}{
		{"password", true},
		{"Password", true},
		{"DATABASE_PASSWORD", true},
		{"token", true},
		{"auth_token", true},
		{"apiKey", true},
		{"api_key", true},
		{"api-key", true},
		{"secret", true},
		{"client_secret", true},
		{"privateKey", true},
		{"private_key", true},
		{"certificate", true},
		{"bearer", true},
		{"authorization", true},
		{"credential", true},
		{"name", false},
		{"namespace", false},
		{"replicas", false},
		{"image", false},
		{"port", false},
		{"log_level", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := isSensitiveKey(strings.ToLower(tc.key))
			if result != tc.sensitive {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tc.key, result, tc.sensitive)
			}
		})
	}
}

func TestSanitizeOutput_PreservesStructure(t *testing.T) {
	// Complex K8s resource
	input := `{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {
			"name": "web-app",
			"namespace": "production",
			"labels": {
				"app": "web"
			}
		},
		"spec": {
			"replicas": 3,
			"selector": {
				"matchLabels": {
					"app": "web"
				}
			}
		}
	}`

	result := SanitizeOutput(input)

	// Verify structure is preserved
	var original, sanitized map[string]interface{}
	if err := json.Unmarshal([]byte(input), &original); err != nil {
		t.Fatalf("Failed to parse original: %v", err)
	}
	if err := json.Unmarshal([]byte(result), &sanitized); err != nil {
		t.Fatalf("Failed to parse sanitized: %v", err)
	}

	// Check that non-sensitive fields are identical
	if sanitized["apiVersion"] != original["apiVersion"] {
		t.Error("apiVersion should not be modified")
	}
	if sanitized["kind"] != original["kind"] {
		t.Error("kind should not be modified")
	}

	metadata := sanitized["metadata"].(map[string]interface{})
	if metadata["name"] != "web-app" {
		t.Error("name should not be modified")
	}
	if metadata["namespace"] != "production" {
		t.Error("namespace should not be modified")
	}
}
