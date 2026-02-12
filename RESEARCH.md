# 🔬 SniffOps 기술 조사

> MCP 프로토콜 + K8s MCP 서버 기존 구현체 분석
> 작성일: 2026-02-11

---

## 1. MCP (Model Context Protocol) 개요

### MCP란?
- Anthropic이 만든 **LLM과 외부 도구 간 통신 표준 프로토콜**
- JSON-RPC 2.0 기반
- LLM이 외부 도구(Tool)를 호출하고 결과를 받을 수 있게 해줌
- Claude Code, Claude Desktop 등에서 공식 지원

### 핵심 개념
```
[LLM Client] ←(JSON-RPC)→ [MCP Server] ←→ [외부 시스템]
 (Claude Code)                (우리가 만들 것)     (K8s API)
```

- **Server**: Tool을 제공하는 쪽 (SniffOps)
- **Client**: Tool을 호출하는 쪽 (Claude Code)
- **Tool**: 서버가 제공하는 기능 단위 (예: kubectl_get, kubectl_apply)
- **Transport**: 통신 방식 (stdio, SSE, HTTP)

### Transport 종류
| Transport | 설명 | Claude Code 지원 |
|-----------|------|:---:|
| **stdio** | stdin/stdout 통신. 로컬 프로세스 | ✅ (기본) |
| **SSE** | Server-Sent Events. HTTP 기반 | ✅ |
| **Streamable HTTP** | 최신 HTTP 기반 | ✅ |

**SniffOps MVP는 stdio 사용 추천** — 가장 간단하고 Claude Code에서 바로 동작

### Claude Code에서 MCP 서버 등록 방법
```bash
# Go 바이너리일 경우
claude mcp add sniffops -- /path/to/sniffops

# 또는 설치 후
claude mcp add sniffops -- sniffops serve
```

설정 파일 (`~/.claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "sniffops": {
      "command": "sniffops",
      "args": ["serve"]
    }
  }
}
```

---

## 2. Go MCP SDK

### 공식 SDK: `modelcontextprotocol/go-sdk`
- **GitHub**: https://github.com/modelcontextprotocol/go-sdk
- **관리**: Anthropic + Google 공동 유지보수
- **MCP 스펙 지원**: 2024-11-05 ~ 2025-06-18 (최신)
- **안정성**: v1.2.0+ (프로덕션 사용 가능)

### 비공식 SDK: `mark3labs/mcp-go`
- **GitHub**: https://github.com/mark3labs/mcp-go
- 초기에 많이 사용됐지만, 공식 SDK 나온 이후 공식 쪽 추천

### ✅ 결론: 공식 SDK (`modelcontextprotocol/go-sdk`) 사용

### 서버 구현 기본 구조

```go
package main

import (
    "context"
    "log"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool 입력 정의
type KubectlGetInput struct {
    Resource  string `json:"resource" jsonschema:"K8s resource type (pod, deployment, etc)"`
    Namespace string `json:"namespace" jsonschema:"K8s namespace"`
    Name      string `json:"name" jsonschema:"resource name (optional)"`
}

// Tool 출력 정의
type KubectlGetOutput struct {
    Result string `json:"result"`
}

// Tool 핸들러
func KubectlGet(ctx context.Context, req *mcp.CallToolRequest, input KubectlGetInput) (
    *mcp.CallToolResult, KubectlGetOutput, error,
) {
    // 1. trace 기록 시작
    // 2. K8s API 호출
    // 3. 결과 + trace 저장
    // 4. 결과 반환
    return nil, KubectlGetOutput{Result: "..."}, nil
}

func main() {
    server := mcp.NewServer(
        &mcp.Implementation{Name: "sniffops", Version: "v0.1.0"}, nil,
    )

    // Tool 등록
    mcp.AddTool(server, &mcp.Tool{
        Name:        "kubectl_get",
        Description: "Get Kubernetes resources",
    }, KubectlGet)

    // stdio로 실행
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

### 주요 패키지
| 패키지 | 용도 |
|--------|------|
| `mcp` | Server/Client, Tool, Transport 핵심 API |
| `jsonrpc` | 커스텀 Transport 구현 시 |
| `auth` | OAuth 지원 (MVP에서는 불필요) |

---

## 3. 기존 K8s MCP 서버 구현체 분석

현재 5개의 주요 K8s MCP 서버가 존재:

### 3-1. containers/kubernetes-mcp-server ⭐ (가장 참고할 만함)
- **GitHub**: https://github.com/containers/kubernetes-mcp-server
- **언어**: Go
- **특징**:
  - K8s API 직접 호출 (kubectl 래핑 아님)
  - 싱글 네이티브 바이너리 (npm, pip, Docker도 지원)
  - 멀티 클러스터 지원
  - OpenShift 지원
  - OTel 트레이싱/메트릭 내장
  - Claude Code 전용 가이드 있음
- **제공 Tool**: Pod CRUD, Namespace, Events, Helm, 범용 리소스 CRUD
- **SniffOps와의 관계**: 
  - 이 서버가 K8s 명령 실행을 담당
  - SniffOps는 이런 서버의 **액션을 감시/기록**하는 역할
  - 경쟁이 아니라 **보완 관계**

### 3-2. Flux159/mcp-server-kubernetes
- **GitHub**: https://github.com/Flux159/mcp-server-kubernetes
- **언어**: TypeScript
- **특징**:
  - kubectl, helm 명령 래핑 방식
  - SSE + stdio 지원
  - 아키텍처 문서가 잘 되어있음 (참고용)
- **제공 Tool**: kubectl_get, kubectl_apply, kubectl_delete, kubectl_scale, helm 등

### 3-3. Azure/mcp-kubernetes
- **GitHub**: https://github.com/Azure/mcp-kubernetes
- **언어**: Go
- **특징**:
  - Microsoft/Azure 공식
  - 단일 `call_kubectl` 도구로 모든 명령 처리
  - 심플한 접근

### 3-4. rohitg00/kubectl-mcp-server
- **GitHub**: https://github.com/rohitg00/kubectl-mcp-server
- **언어**: TypeScript + Python
- **특징**:
  - npx 또는 pip으로 설치
  - 브라우저 기반 K8s 조작 지원

### 3-5. alexei-led/k8s-mcp-server
- **GitHub**: https://github.com/alexei-led/k8s-mcp-server
- **언어**: Go (Docker 기반)
- **특징**:
  - Docker 컨테이너 안에서 kubectl, helm, istioctl, argocd 실행
  - 보안 격리 강조

---

## 4. SniffOps 아키텍처 결정 사항

### 접근 방식: "프록시 MCP 서버"

기존 K8s MCP 서버들은 "K8s 명령을 실행"하는 도구.
SniffOps는 **그 위에 얹어서 감시하는 레이어**.

**두 가지 접근:**

#### 접근 A: 독립 MCP 서버 (자체 K8s 명령 실행 + trace)
```
Claude Code ←→ SniffOps MCP Server ←→ K8s API
                      ↓
                 trace 저장
```
- 장점: 완전한 제어, 의존성 없음
- 단점: K8s 명령 실행 로직을 처음부터 구현해야 함

#### 접근 B: 프록시/미들웨어 (기존 MCP 서버를 감싸기)
```
Claude Code ←→ SniffOps (프록시) ←→ 기존 K8s MCP Server ←→ K8s API
                    ↓
               trace 저장
```
- 장점: 기존 서버 재활용, 개발 빠름
- 단점: 의존성 추가, 설정 복잡

#### ✅ 추천: 접근 A (독립 MCP 서버)

이유:
1. `containers/kubernetes-mcp-server`가 Go + K8s API 직접 호출 방식이라 참고 가능
2. K8s client-go 사용하면 kubectl 래핑보다 깔끔
3. 프록시 방식은 설정이 복잡해져서 MVP에 안 맞음
4. trace 수집 로직을 Tool 핸들러 안에 자연스럽게 넣을 수 있음

### K8s API 접근 방식

kubectl을 호출하지 않고, **client-go로 직접 K8s API 통신** 추천:
- `containers/kubernetes-mcp-server`도 이 방식
- 외부 의존성(kubectl 바이너리) 불필요
- 응답 파싱이 구조적
- Go 생태계에서 표준

```go
import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)
```

### MVP Tool 목록 (v0.1)

| Tool 이름 | 동작 | 위험도 |
|-----------|------|:------:|
| `sniff_get` | 리소스 조회 (get, describe) | 🟢 low |
| `sniff_logs` | Pod 로그 조회 | 🟢 low |
| `sniff_apply` | 리소스 생성/수정 | 🟡 medium |
| `sniff_delete` | 리소스 삭제 | 🔴 high |
| `sniff_scale` | 레플리카 수 변경 | 🔴 high |
| `sniff_exec` | Pod 내 명령 실행 | 🔴 high |
| `sniff_traces` | 저장된 trace 조회 (자체 기능) | 🟢 low |
| `sniff_stats` | 사용 통계 조회 (자체 기능) | 🟢 low |

### 프로젝트 디렉토리 구조 (초안)

```
sniffops/
├── cmd/
│   └── sniffops/
│       └── main.go          # 엔트리포인트
├── internal/
│   ├── server/
│   │   └── server.go        # MCP 서버 설정
│   ├── tools/
│   │   ├── get.go           # sniff_get
│   │   ├── logs.go          # sniff_logs
│   │   ├── apply.go         # sniff_apply
│   │   ├── delete.go        # sniff_delete
│   │   ├── scale.go         # sniff_scale
│   │   └── exec.go          # sniff_exec
│   ├── trace/
│   │   ├── recorder.go      # trace 기록
│   │   ├── store.go         # SQLite 저장
│   │   └── sanitizer.go     # 민감 정보 마스킹
│   ├── risk/
│   │   └── evaluator.go     # 위험도 평가
│   ├── k8s/
│   │   └── client.go        # K8s client-go 래퍼
│   └── web/
│       ├── handler.go       # 웹 API 핸들러
│       └── embed.go         # React 빌드 임베드
├── web/                      # React 프론트엔드
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── LICENSE                   # Apache 2.0
```

---

## 5. 핵심 참고 자료

| 자료 | URL | 용도 |
|------|-----|------|
| MCP 공식 Go SDK | https://github.com/modelcontextprotocol/go-sdk | MCP 서버 구현 |
| MCP Go SDK 문서 | https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp | API 레퍼런스 |
| containers/kubernetes-mcp-server | https://github.com/containers/kubernetes-mcp-server | Go K8s MCP 참고 구현 |
| Flux159/mcp-server-kubernetes | https://github.com/Flux159/mcp-server-kubernetes | 아키텍처 참고 |
| Claude Code MCP 가이드 | https://code.claude.com/docs/en/mcp | Claude Code 연동 방법 |
| MCP 서버 빌드 가이드 (Go) | https://navendu.me/posts/mcp-server-go/ | 실전 튜토리얼 |
| client-go 문서 | https://pkg.go.dev/k8s.io/client-go | K8s API 접근 |

---

## 6. 다음 단계

1. [ ] Go 프로젝트 초기화 (`go mod init github.com/sniffops/sniffops`)
2. [ ] MCP 공식 Go SDK 연동 테스트 (Hello World 수준)
3. [ ] client-go로 K8s API 연동 테스트
4. [ ] `sniff_get` Tool 1개 구현 + trace 기록
5. [ ] SQLite 스토리지 구현
6. [ ] Claude Code에서 테스트

---

_이 문서는 기술 조사 결과물이며, 개발 진행에 따라 업데이트됩니다._
_최초 작성: 2026-02-11_
