# MCP Go SDK 심층 분석

**작성일**: 2026-02-12  
**조사자**: Scout (SniffOps Researcher)  
**태스크**: TASK-002

---

## 📋 요약

MCP Go SDK는 Model Context Protocol의 공식 Go 구현체로, Google과의 협업으로 유지 관리되고 있습니다. 2026년 2월 기준 최신 버전은 v1.3.0이며, MCP Spec 2025-11-25까지 지원합니다.

**주요 출처**:
- GitHub: https://github.com/modelcontextprotocol/go-sdk
- Go Packages: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp
- 릴리스 노트: https://github.com/modelcontextprotocol/go-sdk/releases

---

## 1️⃣ 최신 버전 및 릴리스 현황

### 현재 버전
- **최신 stable**: v1.3.0 (2026년 초 릴리스)
- **라이선스**: Apache 2.0 (신규 기여), MIT (기존 코드)
- **유지 관리**: Google과 협업 중

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.3.0

### 버전 호환성 매트릭스

| SDK 버전 | 최신 MCP Spec | 지원하는 모든 Spec |
|---------|--------------|------------------|
| v1.2.0+ | 2025-06-18 | 2025-11-25, 2025-06-18, 2025-03-26, 2024-11-05 |
| v1.0.0 - v1.1.0 | 2025-06-18 | 2025-06-18, 2025-03-26, 2024-11-05 |

**출처**: https://github.com/modelcontextprotocol/go-sdk (README의 Version Compatibility 섹션)

### v1.3.0 주요 변경사항
1. **성능 개선**: Schema caching 추가로 stateless 서버 배포 시나리오에서 성능 대폭 향상
2. **로깅 개선**: ClientOptions에 Logger 추가 (deprecated logger 제거)
3. **버그 수정**: SSE connection, logging race condition 등 수정
4. **의존성 업데이트**: jsonschema v0.4.2로 업그레이드

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.3.0

---

## 2️⃣ Tool 등록 방법

### 기본 패턴: 제네릭 `AddTool` 함수

MCP Go SDK는 **타입 안전한 제네릭 함수**를 제공하여 Tool 등록을 단순화합니다.

```go
package main

import (
    "context"
    "log"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input 구조체 정의 (JSON Schema가 자동 생성됨)
type Input struct {
    Name string `json:"name" jsonschema:"the name of the person to greet"`
}

// Output 구조체 정의
type Output struct {
    Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

// Tool Handler 함수
func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
    *mcp.CallToolResult,
    Output,
    error,
) {
    return nil, Output{Greeting: "Hi " + input.Name}, nil
}

func main() {
    // 서버 생성
    server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
    
    // Tool 등록 (스키마 자동 생성)
    mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)
    
    // Stdio Transport로 실행
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

**출처**: https://github.com/modelcontextprotocol/go-sdk (README 예제)

### 핵심 특징

1. **자동 스키마 생성**: `jsonschema` 태그를 사용하여 입출력 스키마 자동 추론
2. **타입 안전성**: 제네릭을 사용하여 컴파일 타임에 타입 검증
3. **자동 검증**: 입력값이 스키마에 따라 자동 검증됨
4. **출력 스키마 생략 가능**: Output 타입이 `any`인 경우 output schema 생성 안 함

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#AddTool

### 커스텀 스키마 사용

더 복잡한 스키마가 필요한 경우 직접 정의 가능:

```go
import (
    "reflect"
    "github.com/google/jsonschema-go/jsonschema"
)

// 커스텀 타입에 대한 스키마 정의
customSchemas := map[reflect.Type]*jsonschema.Schema{
    reflect.TypeFor[Probability](): {
        Type: "number", 
        Minimum: jsonschema.Ptr(0.0), 
        Maximum: jsonschema.Ptr(1.0),
    },
}

opts := &jsonschema.ForOptions{TypeSchemas: customSchemas}
inputSchema, err := jsonschema.For[WeatherInput](opts)

// Tool 등록 시 커스텀 스키마 사용
mcp.AddTool(server, &mcp.Tool{
    Name: "weather",
    InputSchema: inputSchema,
}, WeatherTool)
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#example-AddTool-ComplexSchema

---

## 3️⃣ 핸들러 패턴 (요청 → 응답 흐름)

### 아키텍처 개요

```
Client                                    Server
  ⇅ (jsonrpc2) ⇅
ClientSession ⇄ Transport ⇄ Transport ⇄ ServerSession
```

- **Client/Server**: 여러 연결을 동시에 처리 가능
- **Session**: Transport 연결 시마다 생성되는 세션 (ClientSession / ServerSession)
- **Transport**: Client와 Server를 연결하는 통신 계층

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#hdr-Clients__servers__and_sessions

### Tool Handler 시그니처

```go
type ToolHandlerFor[In, Out any] func(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input In,
) (*mcp.CallToolResult, Out, error)
```

**파라미터**:
- `ctx`: Context (cancellation, timeout 지원)
- `req`: 요청 메타데이터 (Session, ProgressToken 등)
- `input`: 자동 파싱된 입력 (스키마 검증 완료)

**반환값**:
- `*mcp.CallToolResult`: 결과 메타데이터 (nil 가능)
- `Out`: 출력 데이터 (자동 직렬화)
- `error`: 에러 발생 시 반환

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#ToolHandlerFor

### Progress Notification 예제

```go
func MakeProgress(ctx context.Context, req *mcp.CallToolRequest, _ any) (
    *mcp.CallToolResult, 
    any, 
    error,
) {
    if token := req.Params.GetProgressToken(); token != nil {
        for i := range 3 {
            params := &mcp.ProgressNotificationParams{
                Message: "frobbing widgets",
                ProgressToken: token,
                Progress: float64(i),
                Total: 2,
            }
            // Progress 알림 전송
            req.Session.NotifyProgress(ctx, params)
        }
    }
    return &mcp.CallToolResult{}, nil, nil
}
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#example-package-Progress

### Cancellation 지원

Context를 통한 취소 전파:

```go
ctx, cancel := context.WithCancel(context.Background())
go func() {
    _, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "slow"})
}()

// 클라이언트에서 취소
cancel()

// 서버 핸들러에서 취소 감지
func SlowTool(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
    select {
    case <-time.After(5 * time.Second):
        return &mcp.CallToolResult{}, nil, nil
    case <-ctx.Done():
        // Context가 취소됨
        return nil, nil, ctx.Err()
    }
}
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#example-package-Cancellation

---

## 4️⃣ 에러 처리 패턴

### 표준 에러 코드

SDK는 MCP-specific 에러 코드를 정의합니다:

```go
const (
    CodeResourceNotFound       = -32002
    CodeURLElicitationRequired = -32042
)
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#pkg-constants

### 에러 생성 헬퍼 함수

```go
// Resource not found 에러
err := mcp.ResourceNotFoundError(uri)

// URL elicitation required 에러
err := mcp.URLElicitationRequiredError(elicitations)
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#ResourceNotFoundError

### JSON-RPC 에러 처리

`jsonrpc` 패키지는 JSON-RPC 2.0 에러를 처리합니다:

```go
import "github.com/modelcontextprotocol/go-sdk/jsonrpc"

// 에러 생성
err := &jsonrpc.Error{
    Code:    -32602,
    Message: "Invalid params",
    Data:    map[string]any{"field": "name"},
}
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/jsonrpc

### CallToolResult 에러 처리

Tool 실행 결과에도 에러를 포함할 수 있습니다:

```go
result, err := session.CallTool(ctx, params)
if err != nil {
    // 통신/프로토콜 에러
    log.Fatal(err)
}

if result.IsError {
    // Tool 실행 에러
    log.Println("Tool failed")
}

// v1.3.0+ GetError/SetError 메서드
if err := result.GetError(); err != nil {
    log.Printf("Tool error: %v", err)
}
```

**출처**: 
- https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#CallToolResult
- https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.3.0 (GetError/SetError 추가)

### 연결 에러

```go
var ErrConnectionClosed error

// 연결이 닫혔거나 닫히는 중일 때 반환됨
if errors.Is(err, mcp.ErrConnectionClosed) {
    // Handle connection closed
}
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#pkg-variables

---

## 5️⃣ Transport 종류

MCP Go SDK는 다양한 Transport를 지원합니다.

### ✅ 지원되는 Transport

#### 1. **StdioTransport**
- **용도**: 로컬 프로세스 간 통신 (stdin/stdout)
- **사용 사례**: CLI 도구, 로컬 MCP 서버

```go
server := mcp.NewServer(&mcp.Implementation{Name: "server"}, nil)
err := server.Run(ctx, &mcp.StdioTransport{})
```

**출처**: https://github.com/modelcontextprotocol/go-sdk (README)

#### 2. **CommandTransport**
- **용도**: 외부 명령 실행 후 stdin/stdout 연결
- **사용 사례**: 클라이언트가 서버 프로세스를 spawn

```go
transport := &mcp.CommandTransport{
    Command: exec.Command("myserver"),
}
session, err := client.Connect(ctx, transport, nil)
```

**출처**: https://github.com/modelcontextprotocol/go-sdk (README)

#### 3. **SSEClientTransport / SSEHandler**
- **용도**: HTTP Server-Sent Events (단방향 푸시)
- **사용 사례**: HTTP 기반 클라이언트-서버 통신 (deprecated, Streamable 권장)
- **특징**: 클라이언트는 GET으로 이벤트 수신, POST로 메시지 전송

```go
// 서버
handler := mcp.NewSSEHandler(getServer, &mcp.SSEOptions{})
http.Handle("/sse", handler)

// 클라이언트
transport := &mcp.SSEClientTransport{
    Endpoint: "https://example.com/sse",
}
```

**출처**: 
- https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#SSEHandler
- https://github.com/orgs/modelcontextprotocol/discussions/364

#### 4. **StreamableHTTPHandler / StreamableClientTransport** ⭐️ 권장
- **용도**: HTTP 기반 양방향 스트리밍 (최신 MCP spec)
- **사용 사례**: 프로덕션 HTTP 서버, 멀티 클라이언트 지원
- **특징**: 
  - Session resumption 지원 (EventStore 사용)
  - SessionTimeout 설정 가능
  - Middleware 지원

```go
// 서버
handler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
    EventStore:     mcp.NewMemoryEventStore(nil),
    SessionTimeout: 30 * time.Minute,
})
http.Handle("/mcp", handler)

// 클라이언트
transport := &mcp.StreamableClientTransport{
    Endpoint: "https://example.com/mcp",
}
```

**출처**: 
- https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#StreamableHTTPHandler
- https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.1.0 (EventStore, SessionTimeout 추가)

#### 5. **InMemoryTransport**
- **용도**: 테스트 및 디버깅
- **사용 사례**: 단위 테스트, in-process 통신

```go
t1, t2 := mcp.NewInMemoryTransports()
serverSession, _ := server.Connect(ctx, t1, nil)
clientSession, _ := client.Connect(ctx, t2, nil)
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#NewInMemoryTransports

#### 6. **IOTransport**
- **용도**: 일반적인 io.ReadCloser / io.WriteCloser 연결
- **사용 사례**: 커스텀 Transport 구현

```go
transport := &mcp.IOTransport{
    Reader: reader,
    Writer: writer,
}
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.1.0

### ❌ 미지원 Transport

- **WebSocket**: 아직 미지원 (Issue #652에서 논의 중)

**출처**: https://github.com/modelcontextprotocol/go-sdk/issues/652

### Transport 비교표

| Transport | 양방향 | HTTP | 프로덕션 권장 | 세션 재개 | 사용 사례 |
|-----------|--------|------|---------------|-----------|-----------|
| Stdio | ✅ | ❌ | ✅ | ❌ | CLI, 로컬 프로세스 |
| Command | ✅ | ❌ | ✅ | ❌ | 클라이언트가 서버 spawn |
| SSE | 부분적 | ✅ | ❌ (deprecated) | ❌ | HTTP 단방향 |
| Streamable HTTP | ✅ | ✅ | ✅ | ✅ | 프로덕션 HTTP 서버 |
| InMemory | ✅ | ❌ | ❌ | ❌ | 테스트 |
| IO | ✅ | ❌ | ✅ | ❌ | 커스텀 구현 |

**출처**: 종합 분석

---

## 6️⃣ 실제 사용 사례 / 예제 프로젝트

### 공식 예제

MCP Go SDK는 `examples/` 디렉토리에 다양한 예제를 제공합니다:

```
examples/
├── client/          # 클라이언트 예제
├── server/          # 서버 예제
│   ├── conformance/ # 적합성 테스트 서버
│   └── ...
└── ...
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/tree/main/examples

### pkg.go.dev 예제 목록

공식 문서에서 제공하는 예제:

1. **Cancellation**: Context 취소 전파
2. **Elicitation**: 클라이언트 elicitation 처리
3. **Lifecycle**: 세션 초기화 및 종료
4. **Logging**: 로깅 메시지 핸들링
5. **Progress**: Progress notification
6. **Prompts**: Prompt 등록 및 사용
7. **Resources**: Resource 및 ResourceTemplate
8. **Roots**: Root 관리
9. **Sampling**: CreateMessage 샘플링
10. **ComplexSchema**: 복잡한 스키마 정의
11. **CustomMarshalling**: 커스텀 JSON 마샬링
12. **LoggingTransport**: Transport 로깅
13. **SSEHandler**: SSE HTTP 서버
14. **StreamableHTTPHandler**: Streamable HTTP 서버 및 Middleware

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#pkg-examples

### 적합성 테스트 서버

MCP 적합성 테스트를 위한 참조 구현:

```bash
# 적합성 테스트 실행
./scripts/conformance.sh
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.2.0

### 대체 SDK 비교

공식 SDK 외에도 서드파티 Go SDK가 존재:

1. **mark3labs/mcp-go**: 초기 커뮤니티 구현 (Ed Zynda 작성)
2. **metoro-io/mcp-golang**: 대체 구현
3. **ThinkInAIXYZ/go-mcp**: 또 다른 구현

**공식 SDK가 이들로부터 영감을 받았으며, README에서 감사를 표함.**

**출처**: https://github.com/modelcontextprotocol/go-sdk (Acknowledgements 섹션)

---

## 7️⃣ SniffOps에서 활용 시 주의사항

### ✅ 권장 사항

#### 1. **Transport 선택**
- **CLI 기반**: `StdioTransport` 사용 (stdin/stdout)
- **HTTP 서버**: `StreamableHTTPHandler` 사용 (최신 spec)
- **테스트**: `InMemoryTransport` 사용

**이유**: SSE는 deprecated이며, StreamableHTTP가 최신 MCP spec을 완전히 지원합니다.

#### 2. **Tool 등록**
```go
// ✅ 권장: 제네릭 AddTool 사용
mcp.AddTool(server, &mcp.Tool{Name: "sniff"}, SniffHandler)

// ❌ 비권장: 저수준 Server.AddTool 사용
server.AddTool(&mcp.Tool{...}, rawHandler)
```

**이유**: 자동 스키마 생성, 검증, 타입 안전성 보장.

#### 3. **에러 처리**
```go
func SniffHandler(ctx context.Context, req *mcp.CallToolRequest, input Input) (
    *mcp.CallToolResult,
    Output,
    error,
) {
    // Context 취소 확인
    select {
    case <-ctx.Done():
        return nil, Output{}, ctx.Err()
    default:
    }
    
    // 비즈니스 로직
    result, err := doSniff(input)
    if err != nil {
        // Tool-level 에러는 error로 반환
        return nil, Output{}, err
    }
    
    return &mcp.CallToolResult{}, result, nil
}
```

#### 4. **Schema Caching (v1.3.0+)**
- Schema 생성이 반복적일 경우 `mcp.NewSchemaCache()` 사용
- Stateless 서버 환경에서 성능 대폭 향상

```go
cache := mcp.NewSchemaCache()
// Cache는 자동으로 사용됨 (내부 구현)
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.3.0

#### 5. **Logging**
```go
// 서버 사이드 로깅
server := mcp.NewServer(&mcp.Implementation{...}, &mcp.ServerOptions{
    Logger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
})

// 클라이언트에게 로그 전송
logger := slog.New(mcp.NewLoggingHandler(serverSession, nil))
logger.Info("Processing packet", "size", packetSize)
```

**출처**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#example-package-Logging

### ⚠️ 주의사항

#### 1. **Tool Name 검증**
Tool 이름은 반드시 regex 패턴 준수:
```
^[a-zA-Z0-9_]{1,64}$
```

**위반 시 Claude 등의 클라이언트에서 에러 발생.**

**출처**: https://github.com/modelcontextprotocol/go-sdk/issues/169

#### 2. **WebSocket 미지원**
- 현재 WebSocket Transport는 미지원
- 필요 시 대체 SDK 검토 또는 직접 구현 필요

**출처**: https://github.com/modelcontextprotocol/go-sdk/issues/652

#### 3. **Session Resumption**
- StreamableHTTP의 기본값이 v1.1.0부터 변경됨:
  - **이전**: 기본 in-memory EventStore 사용
  - **현재**: 기본적으로 비활성화

**Resumption이 필요한 경우 명시적으로 EventStore 설정:**

```go
handler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
    EventStore: mcp.NewMemoryEventStore(&mcp.MemoryEventStoreOptions{
        MaxBytes: 10 * 1024 * 1024, // 10MB
    }),
})
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.1.0

#### 4. **CRLF 처리 (Windows)**
- v1.2.0에서 Windows CRLF 처리 버그 수정됨
- Windows 환경에서는 v1.2.0 이상 사용 권장

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.2.0

#### 5. **OAuth 2.0**
- 서버 사이드 OAuth는 `auth` 패키지로 지원
- **클라이언트 사이드 OAuth는 실험적 기능** (`mcp_go_client_oauth` 빌드 태그 필요)

```bash
go build -tags=mcp_go_client_oauth
```

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.1.0

#### 6. **의존성 관리**
공식 SDK는 다음 주요 의존성 사용:
- `github.com/google/jsonschema-go` (v0.4.2+)
- `golang.org/x/exp` (jsonrpc2)

**의존성 충돌 주의.**

**출처**: https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.3.0

### 🎯 SniffOps 적용 시나리오

#### 시나리오 1: CLI 기반 패킷 분석 도구
```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "sniffops",
    Version: "v0.1.0",
}, nil)

type SniffInput struct {
    Interface string `json:"interface" jsonschema:"network interface to sniff"`
    Filter    string `json:"filter,omitempty" jsonschema:"BPF filter expression"`
}

type SniffOutput struct {
    Packets []Packet `json:"packets"`
    Stats   Stats    `json:"stats"`
}

mcp.AddTool(server, &mcp.Tool{
    Name:        "sniff",
    Description: "Capture network packets",
}, SniffHandler)

server.Run(context.Background(), &mcp.StdioTransport{})
```

#### 시나리오 2: HTTP API 서버
```go
handler := mcp.NewStreamableHTTPHandler(
    func(r *http.Request) *mcp.Server {
        // 인증 로직
        return server
    },
    &mcp.StreamableHTTPOptions{
        Logger:         logger,
        SessionTimeout: 15 * time.Minute,
    },
)

http.Handle("/mcp", handler)
http.ListenAndServe(":8080", nil)
```

---

## 📚 추가 참고 자료

### 공식 문서
- **GitHub Repo**: https://github.com/modelcontextprotocol/go-sdk
- **Go Packages**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk
- **MCP Spec**: https://modelcontextprotocol.io/specification/
- **Feature Docs**: https://github.com/modelcontextprotocol/go-sdk/tree/main/docs

### 커뮤니티
- **Discussions**: https://github.com/orgs/modelcontextprotocol/discussions
- **Design Discussion**: https://github.com/orgs/modelcontextprotocol/discussions/364
- **Issues**: https://github.com/modelcontextprotocol/go-sdk/issues

### 대체 SDK
- **mark3labs/mcp-go**: https://github.com/mark3labs/mcp-go
- **riza-io/mcp-go**: https://github.com/riza-io/mcp-go
- **metoro-io/mcp-golang**: https://github.com/metoro-io/mcp-golang

---

## ✍️ 확인된 사실 vs 추측

### ✅ 확인된 사실
- 최신 버전: v1.3.0 (릴리스 노트로 확인)
- Tool 등록: 제네릭 `AddTool` 함수 사용 (공식 문서)
- Transport 종류: Stdio, Command, SSE, StreamableHTTP, InMemory, IO (README 및 pkg.go.dev)
- 에러 처리: jsonrpc.Error, ResourceNotFoundError 등 (pkg.go.dev API 문서)
- Schema caching: v1.3.0에서 성능 개선 (릴리스 노트)
- WebSocket 미지원 (Issue #652)

### 🤔 추측
- SniffOps 구체적 적용 시나리오는 SniffOps 요구사항에 따라 조정 필요
- 프로덕션 환경의 구체적 성능 수치는 실측 필요

---

**조사 완료일**: 2026-02-12  
**다음 단계**: 실제 SniffOps 코드베이스에 통합하여 PoC 구현 권장
