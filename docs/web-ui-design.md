# SniffOps Web UI 설계 (MVP)

> **작성일**: 2026-02-14  
> **목표**: 라즈베리파이 ARM64에서 동작하는 경량 Web UI Dashboard  
> **원칙**: 싱글 바이너리, 가볍게, 핵심 기능만

---

## 1. 아키텍처 설계

### 1.1 전체 구조

```
┌─────────────────────────────────────────────────────────┐
│  sniffops (single binary)                               │
│                                                         │
│  ┌─────────────┐        ┌──────────────────┐          │
│  │ MCP Server  │        │  HTTP API Server │          │
│  │ (stdio)     │        │  (port 3000)     │          │
│  │             │        │                  │          │
│  │ cmd/main.go │        │  internal/web/   │          │
│  │ serve       │        │    - api.go      │          │
│  └─────────────┘        │    - embed.go    │          │
│                         └─────────┬────────┘          │
│                                   │                    │
│                    ┌──────────────┴──────────────┐    │
│                    │  internal/trace/store.go    │    │
│                    │  SQLite DB                  │    │
│                    └─────────────────────────────┘    │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │  Embedded Frontend (web/dist/)                   │ │
│  │  - HTML/CSS/JS (Vanilla or Preact)              │ │
│  │  - Bundled via Go embed                         │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 1.2 선택한 접근 방식: **내장 HTTP 서버 (Single Binary)**

**이유**:
- ✅ 싱글 바이너리 유지 (배포 간편)
- ✅ `go:embed`로 프론트엔드 번들링 → 추가 파일 불필요
- ✅ MCP 서버와 Web 서버는 별도 프로세스로 실행 (독립적)
  - `sniffops serve` → stdio MCP 서버 (Claude Code용)
  - `sniffops web` → HTTP API + Web UI (사용자용)

**구현 방식**:
```go
// cmd/sniffops/main.go에 이미 구조 존재 (TODO 상태)
func runWeb() error {
    // 1. SQLite DB 연결
    // 2. HTTP API 핸들러 등록
    // 3. Embedded React/Preact UI 서빙
}
```

---

## 2. API 엔드포인트 설계

### 2.1 REST API 스펙

| 메소드 | 경로 | 설명 | 쿼리 파라미터 |
|--------|------|------|--------------|
| `GET` | `/api/traces` | 트레이스 목록 조회 (필터링/페이징) | `?tool=sniff_get&namespace=prod&risk=high&limit=50&offset=0&start=unix_ms&end=unix_ms` |
| `GET` | `/api/traces/:id` | 특정 트레이스 상세 조회 | - |
| `GET` | `/api/stats` | 통계 데이터 (위험도 분포, 도구별 사용량, 시간대별 트렌드) | `?period=24h` |
| `GET` | `/api/namespaces` | 네임스페이스 목록 (필터 자동완성용) | - |
| `GET` | `/api/tools` | 도구 목록 (필터 자동완성용) | - |
| `GET` | `/` | Web UI (정적 파일 서빙) | - |

### 2.2 응답 예시

#### `GET /api/traces`
```json
{
  "traces": [
    {
      "id": "trace-abc123",
      "session_id": "session-xyz",
      "timestamp": 1708059600000,
      "tool_name": "sniff_get",
      "command": "kubectl get pods -n production",
      "namespace": "production",
      "resource_kind": "pod",
      "target_resource": "nginx-7d9c8f",
      "risk_level": "medium",
      "risk_reason": "Critical namespace: production",
      "result": "success",
      "latency_ms": 245
    }
  ],
  "total": 156,
  "limit": 50,
  "offset": 0
}
```

#### `GET /api/stats`
```json
{
  "risk_distribution": {
    "critical": 3,
    "high": 12,
    "medium": 45,
    "low": 96
  },
  "tool_usage": {
    "sniff_get": 78,
    "sniff_logs": 34,
    "sniff_apply": 12,
    "sniff_delete": 3
  },
  "timeline": [
    {"hour": "2026-02-14T09:00:00Z", "count": 8},
    {"hour": "2026-02-14T10:00:00Z", "count": 15}
  ],
  "total_operations": 156,
  "total_cost_estimate": 0.0234
}
```

---

## 3. 프론트엔드 설계

### 3.1 기술 스택 (가벼운 옵션)

**선택 1: Vanilla JS + Tailwind CSS** (최경량)
- ✅ 번들러 불필요 (단일 HTML + CDN)
- ✅ 빌드 스텝 최소화
- ❌ 복잡한 상태 관리 어려움

**선택 2: React + Tailwind CSS + shadcn/ui** (채택)
- ✅ shadcn/ui 컴포넌트 그대로 사용 가능
- ✅ 프로덕션급 UI 퀄리티
- ✅ Vite로 빌드 → `web/dist/` 디렉터리 생성 → Go embed
- ✅ ARM64에서도 빌드 가능

**최종 선택**: **React + TypeScript + Tailwind + shadcn/ui + Vite**

### 3.2 디렉터리 구조

```
web/
├── src/
│   ├── main.jsx              # 엔트리포인트
│   ├── App.jsx               # 메인 앱 컴포넌트
│   ├── components/
│   │   ├── TraceTimeline.jsx # 타임라인 뷰
│   │   ├── RiskDashboard.jsx # 위험도 대시보드
│   │   ├── FilterBar.jsx     # 필터 UI
│   │   ├── TraceDetail.jsx   # 트레이스 상세 모달
│   │   └── Stats.jsx         # 통계 위젯
│   ├── api.js                # API 클라이언트 (fetch wrapper)
│   └── utils.js              # 유틸리티 (시간 포맷, 색상 매핑)
├── index.html
├── vite.config.js
└── package.json

dist/                         # Vite 빌드 출력
└── assets/
    ├── index-abc123.js
    └── index-def456.css
```

### 3.3 Go Embed 통합

```go
// internal/web/embed.go
package web

import "embed"

//go:embed dist/*
var DistFS embed.FS
```

```go
// cmd/sniffops/main.go
import "github.com/sniffops/sniffops/internal/web"

func runWeb() error {
    // ...
    http.Handle("/", http.FileServer(http.FS(web.DistFS)))
}
```

---

## 4. 핵심 화면 설계

### 4.1 메인 레이아웃

```
┌───────────────────────────────────────────────────────┐
│  🔍 SniffOps Dashboard                    [Period ▼]  │
├───────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────┐ │
│  │  Risk Distribution                              │ │
│  │  🔴 Critical: 3  🟠 High: 12  🟡 Med: 45  🟢 Low│ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  Filters                                        │ │
│  │  [Tool ▼] [Namespace ▼] [Risk ▼] [Time Range] │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  Trace Timeline (시간 역순)                     │ │
│  │  ────────────────────────────────────────────── │ │
│  │  10:45:23  🟡 sniff_apply  production/nginx    │ │
│  │            "Apply deployment config"            │ │
│  │  ────────────────────────────────────────────── │ │
│  │  10:42:15  🔴 sniff_delete kube-system/pod-x   │ │
│  │            "Critical namespace deletion"        │ │
│  │  ────────────────────────────────────────────── │ │
│  │  [Load More]                                    │ │
│  └─────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────┘
```

### 4.2 화면별 상세 설계

#### **A. 위험도 대시보드** (RiskDashboard.jsx)
- **목적**: 전체 운영의 위험도 분포를 한눈에 파악
- **컴포넌트**:
  - 카드 4개 (Critical, High, Medium, Low)
  - 각 카드: 숫자 + 아이콘 + 클릭 시 필터링
  - 색상: `critical=red-600`, `high=orange-500`, `medium=yellow-500`, `low=green-500`
  
**데이터 소스**: `GET /api/stats` → `risk_distribution`

#### **B. 트레이스 타임라인** (TraceTimeline.jsx)
- **목적**: 시간순 MCP 호출 기록 (최신순)
- **표시 정보** (각 행):
  - 시간 (HH:MM:SS)
  - 위험도 뱃지 (🔴🟠🟡🟢)
  - 도구 이름 (`sniff_get`)
  - 타겟 리소스 (`production/nginx`)
  - 명령어 요약 (20자 truncate)
  - 결과 상태 (✅ success / ❌ error)
- **인터랙션**:
  - 행 클릭 → 상세 모달 (`TraceDetail`)
  - 무한 스크롤 or "Load More" 버튼

**데이터 소스**: `GET /api/traces?limit=50&offset=0`

#### **C. 필터 바** (FilterBar.jsx)
- **필터 옵션**:
  - Tool: 드롭다운 (`GET /api/tools`)
  - Namespace: 드롭다운 (`GET /api/namespaces`)
  - Risk Level: 체크박스 (Critical/High/Medium/Low)
  - Time Range: Date picker (Last 24h / 7d / 30d / Custom)
- **동작**: 필터 변경 시 URL 쿼리 업데이트 + API 재호출

#### **D. 트레이스 상세 모달** (TraceDetail.jsx)
- **표시 항목**:
  - Session ID
  - Timestamp (ISO 8601)
  - Tool Name
  - Full Command (`kubectl ...`)
  - Target Resource
  - Namespace
  - Resource Kind
  - Risk Level + Reason
  - Result (성공/실패)
  - Output (YAML/JSON pretty print)
  - Error Message (있으면)
  - Latency (ms)
  - Tokens (input/output)
  - Cost Estimate

**UI**: 모달 오버레이 (ESC로 닫기)

#### **E. 통계 위젯** (Stats.jsx)
- **표시 정보**:
  - Total Operations
  - Total Cost
  - Most Used Tool (bar chart)
  - Hourly Trend (simple line chart or sparkline)

**데이터 소스**: `GET /api/stats`

---

## 5. 구현 태스크 분할

### 5.1 Gopher (백엔드 담당)

#### **태스크 1: HTTP API 서버 구현** (`internal/web/api.go`)
- [ ] HTTP 서버 초기화 (Gin or net/http)
- [ ] CORS 설정 (개발 시 localhost:5173 허용)
- [ ] 엔드포인트 구현:
  - [ ] `GET /api/traces` (필터링 + 페이징)
  - [ ] `GET /api/traces/:id`
  - [ ] `GET /api/stats` (집계 쿼리)
  - [ ] `GET /api/namespaces` (DISTINCT query)
  - [ ] `GET /api/tools` (DISTINCT query)
- [ ] 에러 핸들링 (JSON 에러 응답)

**파일**: `internal/web/api.go`, `internal/web/server.go`

#### **태스크 2: 통계 쿼리 구현** (`internal/trace/stats.go`)
- [ ] 위험도 분포 집계 (`GROUP BY risk_level`)
- [ ] 도구별 사용량 집계 (`GROUP BY tool_name`)
- [ ] 시간대별 트렌드 (hourly buckets)
- [ ] 비용 총합 (`SUM(cost_estimate)`)

**파일**: `internal/trace/stats.go`

#### **태스크 3: Go Embed 통합**
- [ ] `internal/web/embed.go` 생성
- [ ] `web/dist` 디렉터리 embed
- [ ] `cmd/sniffops/main.go`에서 `http.FileServer` 연결
- [ ] 빌드 스크립트에 프론트엔드 빌드 단계 추가

**파일**: `internal/web/embed.go`, `Makefile`

---

### 5.2 Bee DJ (프론트엔드 담당)

#### **태스크 1: 프로젝트 초기화**
- [ ] `web/` 디렉터리에 Vite + Preact 프로젝트 생성
  ```bash
  npm create vite@latest web -- --template preact
  cd web && npm install
  npm install -D tailwindcss postcss autoprefixer
  npx tailwindcss init -p
  ```
- [ ] Tailwind CSS 설정
- [ ] `vite.config.js` 빌드 경로 설정 (`outDir: 'dist'`)

**파일**: `web/package.json`, `web/vite.config.js`, `web/tailwind.config.js`

#### **태스크 2: API 클라이언트 작성**
- [ ] `src/api.js` 생성
  ```js
  export async function fetchTraces(filters) {
    const params = new URLSearchParams(filters);
    const res = await fetch(`/api/traces?${params}`);
    return res.json();
  }
  ```
- [ ] `fetchTraceById`, `fetchStats`, `fetchNamespaces`, `fetchTools` 함수 추가

**파일**: `web/src/api.js`

#### **태스크 3: 컴포넌트 구현**
- [ ] `RiskDashboard.jsx`: 위험도 카드 4개 (critical/high/medium/low)
- [ ] `FilterBar.jsx`: 드롭다운 + 날짜 선택기
- [ ] `TraceTimeline.jsx`: 무한 스크롤 리스트
- [ ] `TraceDetail.jsx`: 모달 (트레이스 상세)
- [ ] `Stats.jsx`: 간단한 통계 위젯

**파일**: `web/src/components/*.jsx`

#### **태스크 4: 메인 앱 구성**
- [ ] `App.jsx`: 레이아웃 + 라우팅 (필요 시)
- [ ] 상태 관리: React hooks (`useState`, `useEffect`)
- [ ] 필터 상태 → URL 쿼리 동기화

**파일**: `web/src/App.jsx`

#### **태스크 5: 빌드 및 테스트**
- [ ] `npm run build` → `dist/` 생성 확인
- [ ] Go 서버와 통합 테스트 (`sniffops web`)
- [ ] 라즈베리파이에서 동작 확인

---

## 6. 제약사항 체크리스트

- [x] **싱글 바이너리**: Go embed 사용
- [x] **라즈베리파이 ARM64**: Go 크로스 컴파일 + Vite 빌드 (Node.js ARM64 지원)
- [x] **가벼운 프레임워크**: Preact (3KB) + Tailwind (purge 후 ~10KB)
- [x] **MVP 수준**: 핵심 3개 화면 (타임라인, 위험도, 통계)

---

## 7. 빌드 프로세스

### 7.1 Makefile 추가

```makefile
.PHONY: build-web build-all

build-web:
	cd web && npm install && npm run build

build-backend:
	go build -o bin/sniffops cmd/sniffops/main.go

build-all: build-web build-backend
	@echo "✅ SniffOps built successfully (backend + frontend)"

clean:
	rm -rf bin/sniffops web/dist web/node_modules
```

### 7.2 릴리즈 프로세스

```bash
# 1. 프론트엔드 빌드
make build-web

# 2. Go 바이너리 빌드 (ARM64)
GOOS=linux GOARCH=arm64 go build -o sniffops cmd/sniffops/main.go

# 3. 실행
./sniffops web --port 3000
```

---

## 8. 다음 단계 (Post-MVP)

1. **실시간 업데이트**: WebSocket or SSE로 새 트레이스 자동 갱신
2. **세션 기반 필터**: 특정 AI 세션의 모든 작업 추적
3. **비용 분석**: 시간대별/도구별 LLM 비용 차트
4. **알림 설정**: Critical 작업 발생 시 Slack/Discord 알림
5. **Export 기능**: 트레이스를 CSV/JSON으로 다운로드
6. **다크 모드**: Tailwind dark variant

---

## 9. 참고 자료

- **MCP 프로토콜**: https://modelcontextprotocol.io
- **Preact 문서**: https://preactjs.com
- **Go embed**: https://pkg.go.dev/embed
- **SQLite CGO-free**: https://gitlab.com/cznic/sqlite

---

**End of Design Document** 🚀
