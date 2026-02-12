# SniffOps Architecture (v0.1 MVP)

> **목표**: AI가 K8s에서 한 모든 행동을 추적·분석하는 Self-hosted O11y 플랫폼
> **원칙**: Simplicity First, YAGNI, 싱글 바이너리
> **기술**: Go + MCP + SQLite + React

---

## 1. System Architecture

### 1.1 Overall Structure

```
┌─────────────────────────────────────────────────────────────────┐
│ 사용자 로컬 머신                                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐                                                │
│  │ Claude Code │                                                │
│  └──────┬──────┘                                                │
│         │ JSON-RPC 2.0 (stdio)                                  │
│         ↓                                                       │
│  ┌──────────────────────────────────┐                          │
│  │   SniffOps MCP Server (Go)       │                          │
│  ├──────────────────────────────────┤                          │
│  │ • Tool Handlers (get/apply/...)  │                          │
│  │ • Trace Recorder                 │                          │
│  │ • Risk Evaluator                 │                          │
│  │ • SQLite Store                   │                          │
│  │ • Embedded Web Server            │                          │
│  └─────┬────────────────────┬───────┘                          │
│        │                    │                                   │
│        │ client-go          │ HTTP :3000                        │
│        ↓                    ↓                                   │
│  ┌──────────┐        ┌─────────────┐                           │
│  │ K8s API  │        │ 웹 브라우저   │                           │
│  │(kubeconf)│        │(React UI)   │                           │
│  └──────────┘        └─────────────┘                           │
│                                                                 │
│  ┌─────────────────────────────────────────┐                   │
│  │ SQLite DB (traces.db)                   │                   │
│  │ • traces 테이블                          │                   │
│  │ • metadata 테이블                        │                   │
│  └─────────────────────────────────────────┘                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Component Overview

| Component | Description | Tech Stack |
|-----------|-------------|------------|
| **MCP Server** | Claude Code ↔ K8s 중계, trace 수집 | Go + MCP SDK |
| **Trace Recorder** | 명령 실행 전후 기록 | Go |
| **Risk Evaluator** | 명령어 위험도 자동 태깅 | Go (룰 기반) |
| **Storage** | Trace 데이터 로컬 저장 | SQLite |
| **Web UI** | 대시보드, 타임라인, 상세 보기 | React + Vite |
| **K8s Client** | K8s API 직접 호출 (kubectl 래핑 안 함) | client-go |

### 1.3 Deployment Model (MVP)

**싱글 바이너리 배포:**
```bash
# 설치
go install github.com/sniffops/sniffops@latest

# Claude Code에 MCP 서버 등록
claude mcp add sniffops -- sniffops serve

# 웹 UI 시작 (별도 터미널)
sniffops web --port 3000
```

**프로세스 구조:**
- `sniffops serve`: stdio로 MCP 서버 실행 (Claude Code가 관리)
- `sniffops web`: HTTP 서버로 웹 UI 제공 (사용자가 직접 실행)
- SQLite DB는 `~/.sniffops/traces.db`에 공유 저장

---

## 2. MCP Server Design

### 2.1 Tool Catalog (v0.1)

| Tool Name | Description | Risk Level | Input | Output |
|-----------|-------------|:----------:|-------|--------|
| `sniff_get` | K8s 리소스 조회 (get, describe) | 🟢 low | resource, namespace, name | YAML/JSON |
| `sniff_logs` | Pod 로그 조회 | 🟢 low | pod, namespace, tail | log text |
| `sniff_apply` | 리소스 생성/수정 (apply) | 🟡 medium | manifest, namespace | apply result |
| `sniff_delete` | 리소스 삭제 | 🔴 high | resource, namespace, name | deletion result |
| `sniff_scale` | 레플리카 수 변경 | 🔴 high | deployment, namespace, replicas | scale result |
| `sniff_exec` | Pod 내 명령 실행 | 🔴 high | pod, namespace, command | command output |
| `sniff_traces` | 저장된 trace 조회 (자체) | 🟢 low | limit, filter | trace list |
| `sniff_stats` | 사용 통계 조회 (자체) | 🟢 low | date_range | stats JSON |

**v0.1에 안 넣는 것:**
- Helm 배포 (v0.2+)
- Custom resource CRUD (v0.2+)
- Multi-cluster 지원 (v0.3+)
- 복잡한 필터링/검색 (v0.2+)

### 2.2 Tool Handler Flow

모든 Tool 핸들러는 동일한 패턴을 따름:

```
┌─────────────────────────────────────────────────────────────┐
│ Tool Handler (예: sniff_get)                                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Input Validation                                        │
│     ├─ namespace, resource, name 검증                       │
│     └─ 필수 파라미터 체크                                    │
│                                                             │
│  2. Trace Recording START                                   │
│     ├─ session_id 생성/재사용                               │
│     ├─ timestamp 기록                                       │
│     ├─ user_intent 파싱 (Claude의 요청 내용)                │
│     └─ risk_level 초기 평가                                 │
│                                                             │
│  3. K8s API Call                                            │
│     ├─ client-go로 K8s API 호출                             │
│     ├─ latency 측정                                         │
│     └─ result 수집                                          │
│                                                             │
│  4. Risk Evaluation                                         │
│     ├─ 명령어 패턴 매칭                                      │
│     ├─ 타겟 리소스 critical 여부 체크                        │
│     └─ 최종 risk_level 결정                                 │
│                                                             │
│  5. Trace Recording END                                     │
│     ├─ result (success/failure) 기록                        │
│     ├─ output 저장 (민감 정보 마스킹)                        │
│     ├─ cost_estimate 계산 (토큰 * 단가)                     │
│     └─ SQLite INSERT                                        │
│                                                             │
│  6. Return to Claude                                        │
│     └─ K8s API result 반환                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 Session Management

**Session ID 생성 규칙:**
- Claude Code는 MCP 서버 프로세스를 재시작하지 않는 한 같은 세션 유지
- SniffOps는 프로세스 시작 시 UUID 생성 → 모든 trace에 동일 session_id 태깅
- 웹 UI에서 "세션별로 그룹화" 가능

```go
// internal/server/server.go
var sessionID string = uuid.New().String()

func getSessionID() string {
    return sessionID
}
```

### 2.4 MCP Server Initialization

```go
// cmd/sniffops/main.go (serve 커맨드)
func runServe() error {
    // 1. SQLite DB 초기화
    db := trace.InitDB("~/.sniffops/traces.db")
    defer db.Close()

    // 2. K8s client 초기화 (kubeconfig 읽기)
    k8sClient := k8s.NewClient()

    // 3. MCP 서버 생성
    server := mcp.NewServer(
        &mcp.Implementation{
            Name:    "sniffops",
            Version: "v0.1.0",
        },
        nil,
    )

    // 4. Tool 등록 (각 핸들러)
    mcp.AddTool(server, &mcp.Tool{
        Name:        "sniff_get",
        Description: "Get Kubernetes resources (pod, deployment, etc)",
    }, tools.NewGetHandler(db, k8sClient))

    // ... (다른 tool들 등록)

    // 5. stdio 통신 시작
    return server.Run(context.Background(), &mcp.StdioTransport{})
}
```

---

## 3. Data Flow: Trace Collection & Storage

### 3.1 End-to-End Flow

```
┌────────────┐   1. 요청                    ┌─────────────────┐
│ Claude Code│───────────────────────────>│SniffOps MCP     │
│            │   "production 네임스페이스의 │                 │
│            │    Pod 목록 보여줘"          │                 │
└────────────┘                             └────────┬────────┘
                                                    │
                                                    │ 2. Trace START
                                                    ↓
                                            ┌───────────────┐
                                            │Trace Recorder │
                                            ├───────────────┤
                                            │ session_id    │
                                            │ timestamp     │
                                            │ user_intent   │
                                            │ risk_level    │
                                            └───────┬───────┘
                                                    │
                                                    │ 3. K8s API 호출
                                                    ↓
                                            ┌───────────────┐
                                            │K8s API Server │
                                            │ GET /pods     │
                                            │ ns=production │
                                            └───────┬───────┘
                                                    │
                                                    │ 4. Result
                                                    ↓
                                            ┌───────────────┐
                                            │Risk Evaluator │
                                            ├───────────────┤
                                            │ command: GET  │
                                            │ ns: production│
                                            │ → low risk    │
                                            └───────┬───────┘
                                                    │
                                                    │ 5. Trace END
                                                    ↓
                                            ┌───────────────┐
                                            │SQLite Store   │
                                            │ INSERT trace  │
                                            └───────────────┘
                                                    ↑
┌────────────┐   6. 결과 반환              │
│ Claude Code│<──────────────────────────────────┘
│ "Pod 3개 발견..." │
└────────────┘

                    7. 사용자 조회
┌────────────┐                             ┌───────────────┐
│ Web UI     │<──────────────────────────>│ SQLite Store  │
│ localhost  │  HTTP GET /api/traces       │               │
│ :3000      │                             │               │
└────────────┘                             └───────────────┘
```

### 3.2 Trace Data Lifecycle

```
1. Receive Tool Call
   ↓
2. Create Trace Record (pending)
   ├─ id: UUID
   ├─ session_id: from process
   ├─ timestamp: now()
   ├─ user_intent: from Claude's prompt
   └─ risk_level: initial eval
   ↓
3. Execute K8s Command
   ├─ measure latency
   └─ capture output
   ↓
4. Finalize Trace Record
   ├─ result: success|failure
   ├─ output: (sanitized)
   ├─ latency_ms: measured
   ├─ cost_estimate: tokens * rate
   └─ risk_level: final eval
   ↓
5. Save to SQLite
   ↓
6. Return to Claude
```

### 3.3 Sanitization (민감 정보 마스킹)

**마스킹 대상:**
- API keys: `apiKey: sk-***`, `OPENAI_API_KEY=***`
- Secrets: `password: ***`, `token: ***`
- URLs with credentials: `https://user:***@host`

**구현:**
```go
// internal/trace/sanitizer.go
func SanitizeOutput(output string) string {
    // Regex 패턴 매칭
    patterns := []struct {
        pattern *regexp.Regexp
        replace string
    }{
        {regexp.MustCompile(`(?i)(api[-_]?key|token|password|secret)\s*[:=]\s*[\w-]+`), "$1: ***"},
        {regexp.MustCompile(`https?://[^:]+:[^@]+@`), "https://***:***@"},
    }
    
    for _, p := range patterns {
        output = p.pattern.ReplaceAllString(output, p.replace)
    }
    return output
}
```

---

## 4. Database Schema (SQLite)

### 4.1 Table: traces

```sql
CREATE TABLE traces (
    -- Identity
    id              TEXT PRIMARY KEY,           -- UUID
    session_id      TEXT NOT NULL,              -- 프로세스별 세션
    timestamp       INTEGER NOT NULL,           -- Unix timestamp (ms)
    
    -- Request Context
    user_intent     TEXT,                       -- Claude에게 사용자가 요청한 내용
    tool_name       TEXT NOT NULL,              -- sniff_get, sniff_apply 등
    
    -- K8s Command Details
    command         TEXT NOT NULL,              -- "kubectl get pods -n prod"
    target_resource TEXT,                       -- "pod/nginx-abc123"
    namespace       TEXT,                       -- "production"
    resource_kind   TEXT,                       -- "pod", "deployment" 등
    
    -- Risk & Security
    risk_level      TEXT NOT NULL,              -- low|medium|high|critical
    risk_reason     TEXT,                       -- "Deletion in production ns"
    
    -- Execution Result
    result          TEXT NOT NULL,              -- success|failure
    output          TEXT,                       -- K8s API 응답 (sanitized)
    error_message   TEXT,                       -- 에러 발생 시
    
    -- Metrics
    latency_ms      INTEGER,                    -- 실행 시간 (ms)
    tokens_input    INTEGER,                    -- LLM 입력 토큰 (추정)
    tokens_output   INTEGER,                    -- LLM 출력 토큰 (추정)
    cost_estimate   REAL,                       -- 비용 추정 (USD)
    
    -- Metadata
    kubeconfig      TEXT,                       -- 사용한 kubeconfig 경로
    cluster_name    TEXT,                       -- K8s 클러스터명 (context)
    
    -- Indexes
    INDEX idx_session_id ON traces(session_id),
    INDEX idx_timestamp ON traces(timestamp DESC),
    INDEX idx_namespace ON traces(namespace),
    INDEX idx_risk_level ON traces(risk_level)
);
```

### 4.2 Table: metadata

```sql
CREATE TABLE metadata (
    key   TEXT PRIMARY KEY,
    value TEXT
);

-- 초기 데이터
INSERT INTO metadata VALUES ('schema_version', '1');
INSERT INTO metadata VALUES ('created_at', datetime('now'));
```

### 4.3 Sample Data

```sql
INSERT INTO traces VALUES (
    '550e8400-e29b-41d4-a716-446655440000',     -- id
    'session-abc123',                           -- session_id
    1707753600000,                              -- timestamp (2024-02-12 20:00:00)
    'production 네임스페이스의 Pod 목록 보여줘', -- user_intent
    'sniff_get',                                -- tool_name
    'kubectl get pods -n production',           -- command
    'pod/*',                                    -- target_resource
    'production',                               -- namespace
    'pod',                                      -- resource_kind
    'low',                                      -- risk_level
    'Read-only operation',                      -- risk_reason
    'success',                                  -- result
    'NAME           READY   STATUS    AGE\nnginx-abc      1/1     Running   5d', -- output
    NULL,                                       -- error_message
    245,                                        -- latency_ms
    150,                                        -- tokens_input
    80,                                         -- tokens_output
    0.0023,                                     -- cost_estimate
    '~/.kube/config',                           -- kubeconfig
    'production-cluster'                        -- cluster_name
);
```

---

## 5. Risk Evaluation Logic

### 5.1 Risk Levels

| Level | Color | Criteria | Examples |
|-------|:-----:|----------|----------|
| **low** | 🟢 | Read-only, 안전한 조회 | get, describe, logs |
| **medium** | 🟡 | 리소스 생성/수정 | apply, patch, port-forward |
| **high** | 🔴 | 리소스 삭제, 스케일 변경 | delete, scale down |
| **critical** | 🔴🔴 | Production 환경 파괴적 작업 | delete in prod ns, scale 0 |

### 5.2 Evaluation Rules

```go
// internal/risk/evaluator.go
type RiskEvaluator struct{}

func (e *RiskEvaluator) Evaluate(ctx EvalContext) RiskLevel {
    // Rule 1: Command Type
    baseRisk := e.getCommandRisk(ctx.ToolName)
    
    // Rule 2: Namespace Criticality
    if e.isCriticalNamespace(ctx.Namespace) {
        baseRisk = e.escalate(baseRisk)
    }
    
    // Rule 3: Resource Count (scale to 0, delete all)
    if ctx.ResourceCount == 0 && ctx.ToolName == "sniff_scale" {
        baseRisk = RiskCritical
    }
    
    return baseRisk
}

func (e *RiskEvaluator) getCommandRisk(tool string) RiskLevel {
    switch tool {
    case "sniff_get", "sniff_logs", "sniff_traces", "sniff_stats":
        return RiskLow
    case "sniff_apply":
        return RiskMedium
    case "sniff_delete", "sniff_scale", "sniff_exec":
        return RiskHigh
    default:
        return RiskMedium
    }
}

func (e *RiskEvaluator) isCriticalNamespace(ns string) bool {
    criticalNS := []string{"production", "prod", "default", "kube-system"}
    for _, c := range criticalNS {
        if ns == c {
            return true
        }
    }
    return false
}

func (e *RiskEvaluator) escalate(level RiskLevel) RiskLevel {
    if level == RiskHigh {
        return RiskCritical
    }
    if level == RiskMedium {
        return RiskHigh
    }
    return level
}
```

### 5.3 Risk Reason Generation

```go
func (e *RiskEvaluator) GetReason(ctx EvalContext, level RiskLevel) string {
    reasons := []string{}
    
    if level >= RiskHigh {
        reasons = append(reasons, fmt.Sprintf("Destructive operation: %s", ctx.ToolName))
    }
    
    if e.isCriticalNamespace(ctx.Namespace) {
        reasons = append(reasons, fmt.Sprintf("Critical namespace: %s", ctx.Namespace))
    }
    
    if ctx.ResourceCount == 0 && ctx.ToolName == "sniff_scale" {
        reasons = append(reasons, "Scaling to 0 replicas")
    }
    
    if len(reasons) == 0 {
        return "Read-only operation"
    }
    
    return strings.Join(reasons, "; ")
}
```

---

## 6. Package Structure (Go)

```
sniffops/
├── cmd/
│   └── sniffops/
│       └── main.go                    # CLI 엔트리포인트 (cobra)
│           ├── serve                  # MCP 서버 시작
│           ├── web                    # 웹 UI 시작
│           └── version                # 버전 출력
│
├── internal/
│   ├── server/
│   │   ├── server.go                  # MCP 서버 초기화 및 설정
│   │   └── session.go                 # 세션 관리
│   │
│   ├── tools/                         # MCP Tool 핸들러들
│   │   ├── get.go                     # sniff_get
│   │   ├── logs.go                    # sniff_logs
│   │   ├── apply.go                   # sniff_apply
│   │   ├── delete.go                  # sniff_delete
│   │   ├── scale.go                   # sniff_scale
│   │   ├── exec.go                    # sniff_exec
│   │   ├── traces.go                  # sniff_traces (조회)
│   │   └── stats.go                   # sniff_stats
│   │
│   ├── trace/
│   │   ├── recorder.go                # Trace 기록 로직
│   │   ├── store.go                   # SQLite CRUD
│   │   ├── sanitizer.go               # 민감 정보 마스킹
│   │   └── models.go                  # Trace 구조체 정의
│   │
│   ├── risk/
│   │   ├── evaluator.go               # 위험도 평가 로직
│   │   └── rules.go                   # 평가 룰 정의
│   │
│   ├── k8s/
│   │   ├── client.go                  # K8s client-go 래퍼
│   │   ├── resources.go               # 리소스 조회/조작
│   │   └── config.go                  # kubeconfig 로딩
│   │
│   └── web/
│       ├── server.go                  # HTTP 서버
│       ├── handler.go                 # API 핸들러 (/api/traces, /api/stats)
│       ├── embed.go                   # React 빌드 임베드 (embed.FS)
│       └── middleware.go              # CORS, logging
│
├── web/                                # React 프론트엔드
│   ├── src/
│   │   ├── App.tsx                    # 메인 앱
│   │   ├── components/
│   │   │   ├── Timeline.tsx           # 타임라인 뷰
│   │   │   ├── TraceDetail.tsx        # trace 상세 모달
│   │   │   └── Stats.tsx              # 통계 대시보드
│   │   ├── api/
│   │   │   └── client.ts              # API 클라이언트
│   │   └── types/
│   │       └── trace.ts               # TypeScript 타입 정의
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
│
├── go.mod
├── go.sum
├── Makefile                            # build, test, install
├── README.md
├── LICENSE                             # Apache 2.0
└── .gitignore
```

### 6.1 Key Interfaces

```go
// internal/trace/models.go
type Trace struct {
    ID            string    `json:"id"`
    SessionID     string    `json:"session_id"`
    Timestamp     int64     `json:"timestamp"`
    UserIntent    string    `json:"user_intent,omitempty"`
    ToolName      string    `json:"tool_name"`
    Command       string    `json:"command"`
    TargetResource string   `json:"target_resource,omitempty"`
    Namespace     string    `json:"namespace,omitempty"`
    ResourceKind  string    `json:"resource_kind,omitempty"`
    RiskLevel     string    `json:"risk_level"`
    RiskReason    string    `json:"risk_reason,omitempty"`
    Result        string    `json:"result"`
    Output        string    `json:"output,omitempty"`
    ErrorMessage  string    `json:"error_message,omitempty"`
    LatencyMs     int       `json:"latency_ms,omitempty"`
    TokensInput   int       `json:"tokens_input,omitempty"`
    TokensOutput  int       `json:"tokens_output,omitempty"`
    CostEstimate  float64   `json:"cost_estimate,omitempty"`
    Kubeconfig    string    `json:"kubeconfig,omitempty"`
    ClusterName   string    `json:"cluster_name,omitempty"`
}

// internal/trace/recorder.go
type Recorder interface {
    Start(ctx context.Context, req RecordRequest) (*Trace, error)
    End(ctx context.Context, trace *Trace, result RecordResult) error
}

// internal/risk/evaluator.go
type Evaluator interface {
    Evaluate(ctx EvalContext) (level string, reason string)
}

// internal/k8s/client.go
type Client interface {
    Get(ctx context.Context, req GetRequest) (string, error)
    Apply(ctx context.Context, manifest string) (string, error)
    Delete(ctx context.Context, req DeleteRequest) error
    Scale(ctx context.Context, req ScaleRequest) error
    Logs(ctx context.Context, req LogsRequest) (string, error)
    Exec(ctx context.Context, req ExecRequest) (string, error)
}
```

---

## 7. Technical Decisions

### 7.1 Go 선택 이유

| 이유 | 설명 |
|------|------|
| **싱글 바이너리** | 웹 UI 임베드, 설치 간편 (`go install` 한 줄) |
| **client-go 네이티브** | kubectl 래핑보다 깔끔하고 안정적 |
| **MCP SDK 공식 지원** | `modelcontextprotocol/go-sdk` 활발히 유지보수 |
| **성능** | stdio 통신, SQLite I/O 모두 빠름 |
| **크로스 플랫폼** | Linux/macOS/Windows 모두 지원 |

### 7.2 MCP 프로토콜 선택 이유

| 이유 | 설명 |
|------|------|
| **Claude Code 네이티브 지원** | 별도 플러그인 불필요 |
| **JSON-RPC 표준** | 디버깅 쉬움, 로깅 명확 |
| **stdio 통신** | 로컬 환경에 최적화 |
| **확장성** | HTTP Transport로 원격 배포 가능 (v0.4+) |

### 7.3 SQLite 선택 이유

| 이유 | 설명 |
|------|------|
| **Zero Configuration** | 별도 DB 서버 불필요 |
| **파일 기반** | 백업/이관 간편 (단일 파일) |
| **충분한 성능** | 개인 사용 기준 수만 건 trace 처리 가능 |
| **임베드 가능** | Go 바이너리에 함께 배포 |
| **경로** | `~/.sniffops/traces.db` (표준 위치) |

**PostgreSQL은 v0.4+에서 옵션으로 제공** (팀 배포 시 중앙 DB 필요)

### 7.4 client-go vs kubectl 래핑

| | client-go (선택) | kubectl exec |
|---|---|---|
| 의존성 | 없음 (Go 라이브러리) | kubectl 바이너리 필요 |
| 성능 | 빠름 (직접 API 호출) | 느림 (프로세스 spawn) |
| 파싱 | 구조체 직접 사용 | stdout 문자열 파싱 |
| 안정성 | K8s API 버전 관리 명확 | kubectl 버전 의존 |
| 코드 품질 | 타입 안전 | 문자열 조작 |

**참고 사례:** `containers/kubernetes-mcp-server`도 client-go 사용

### 7.5 웹 UI 임베드 방식

```go
// internal/web/embed.go
//go:embed all:dist
var webUI embed.FS

func ServeUI() http.Handler {
    fsys := fs.Sub(webUI, "dist")
    return http.FileServer(http.FS(fsys))
}
```

**빌드 프로세스:**
```bash
# 1. React 빌드
cd web && npm run build

# 2. Go 빌드 (embed 포함)
go build -o sniffops ./cmd/sniffops

# 결과: 싱글 바이너리 (웹 UI 포함)
```

### 7.6 Transport: stdio 선택

**MVP에서 stdio를 쓰는 이유:**
- Claude Code가 `command` 방식으로 MCP 서버를 실행하면 자동으로 stdio 통신
- 별도 포트 설정 불필요
- 프로세스 생명주기를 Claude Code가 관리
- 디버깅 시 stdin/stdout 로그 확인 가능

**SSE/HTTP는 v0.4+ (원격 배포 시) 추가 예정**

---

## 8. MVP Boundaries

### 8.1 v0.1에 포함

✅ **Core Features:**
- MCP 서버 (stdio transport)
- Tool: get, logs, apply, delete, scale, exec (6개)
- Trace 수집 및 SQLite 저장
- 위험도 자동 태깅 (룰 기반)
- 민감 정보 마스킹 (API key, secret 패턴)
- 웹 UI: 타임라인 뷰, trace 상세 보기
- CLI: `sniffops serve`, `sniffops web`

✅ **Documentation:**
- README (설치, 사용법)
- Architecture (이 문서)
- API Reference (Tool 명세)

### 8.2 v0.1에 제외

❌ **Not Now:**
- Helm 배포 도구
- Custom Resource CRUD
- Multi-cluster 지원
- 고급 통계/분석
- AI 인사이트
- 알림/노티피케이션
- 사용자 인증
- PostgreSQL 지원
- Export 기능 (JSON/CSV)
- 검색/필터 (v0.2)

### 8.3 Next Steps (v0.2+)

**v0.2 — Analysis:**
- 통계 대시보드 (일별 사용량, 비용, 에러율)
- 검색/필터 (날짜, namespace, risk level)
- Export (JSON, CSV)

**v0.3 — Safety:**
- 위험 명령 실행 전 확인 요청
- 리소스 상태 diff (before/after)
- 커스텀 위험도 룰

**v0.4 — Team/Server:**
- Docker/Helm 배포
- PostgreSQL 지원
- Multi-user + RBAC
- SSE/HTTP transport (원격 MCP)

---

## 9. Development Roadmap

### Phase 1: Foundation (1주)
- [ ] Go 프로젝트 초기화
- [ ] MCP SDK 연동 (hello world)
- [ ] SQLite 스키마 생성 및 CRUD
- [ ] client-go 연동 테스트

### Phase 2: Core Tools (1주)
- [ ] sniff_get 구현 + trace 기록
- [ ] sniff_logs 구현
- [ ] sniff_apply 구현
- [ ] 위험도 평가 로직
- [ ] Claude Code 통합 테스트

### Phase 3: Advanced Tools (3일)
- [ ] sniff_delete, sniff_scale, sniff_exec
- [ ] 민감 정보 마스킹

### Phase 4: Web UI (1주)
- [ ] React 프로젝트 초기화 (Vite)
- [ ] API 엔드포인트 (/api/traces, /api/stats)
- [ ] Timeline 컴포넌트
- [ ] Trace Detail 모달
- [ ] 웹 UI embed 빌드

### Phase 5: Polish (3일)
- [ ] README 작성
- [ ] 설치/사용 가이드
- [ ] 에러 핸들링 개선
- [ ] 테스트 코드 (핵심 로직)

**Total: 약 3주**

---

## 10. Success Metrics (MVP)

**기술적 목표:**
- [ ] Claude Code에서 K8s 명령 10번 실행 → 10개 trace 저장 성공
- [ ] 웹 UI에서 타임라인 조회 < 100ms
- [ ] 싱글 바이너리 크기 < 50MB (웹 UI 포함)
- [ ] 설치 명령어 1줄 (`go install`)
- [ ] Claude Code 설정 1줄 (`claude mcp add`)

**사용성 목표:**
- [ ] 비개발자가 README만 보고 10분 내 설치 가능
- [ ] 위험한 명령(delete in prod) 실행 시 trace에 `critical` 태깅

**확장성 목표:**
- [ ] 1만 개 trace 저장 시 조회 < 500ms
- [ ] SQLite 파일 크기 < 100MB (1만 trace 기준)

---

## 11. Reference Architecture

### 11.1 Comparable Systems

| System | Similarity | Difference |
|--------|-----------|-----------|
| **Langfuse** | Trace 수집/분석 UI | LLM 앱용, 인프라 무관 |
| **kagent** | K8s AI 에이전트 | 자체 에이전트만 지원 |
| **K8s Audit Log** | API 호출 기록 | AI 컨텍스트 없음 |
| **containers/k8s-mcp** | K8s MCP 서버 | O11y 기능 없음 |

**SniffOps는 "K8s MCP 서버" + "LLM O11y" 융합**

### 11.2 Inspiration

- **MCP 서버 구조**: `containers/kubernetes-mcp-server`
- **Trace 데이터 모델**: Langfuse
- **위험도 평가**: AWS IAM Policy Simulator 아이디어
- **웹 UI 디자인**: Grafana 타임라인 뷰

---

_이 문서는 SniffOps v0.1 MVP 아키텍처 명세입니다._
_작성일: 2026-02-12_
_작성자: CTO Architect (Agent)_
