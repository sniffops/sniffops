# 🏢 SniffOps 에이전트 팀 구성 제안서

> OpenClaw 기반 AI 에이전트 팀으로 1인 기업 운영
> 작성일: 2026-02-11

---

## 1. 조사 요약

### 1-1. 가재 컴퍼니 (yuna-openclaw) 분석

[yuna-studio/yuna-openclaw](https://github.com/yuna-studio/yuna-openclaw)는 OpenClaw 기반으로 **13개 에이전트 노드**를 회사처럼 구성한 가장 체계적인 사례:

| 역할 | 기능 |
|------|------|
| **Attendant** (Core OS) | 시스템 운영 총괄, 중앙 관제 |
| **PO** (Product Owner) | 제품 비전, 우선순위 결정 |
| **PM** (Project Manager) | 프로젝트 일정/진행 관리 |
| **BA** (Business Analyst) | 요구사항 분석, 스펙 작성 |
| **DEV** (Developer) | 코드 구현 |
| **QA** (Quality Assurance) | 테스트, 품질 보증 |
| **UX** (UX Designer) | UI/UX 설계 |
| **HR** (Human Resources) | 인사 관리 (에이전트 온보딩/평가) |
| **Legal** | 법무, 컴플라이언스 |
| **Marketing** | 마케팅, 콘텐츠 |
| **CS** (Customer Support) | 고객 대응 |
| + 2 추가 에이전트 | 미확인 |

**핵심 인사이트:**
- 7대 지능 계층 (Core → Business → Task → Chronicle → Governance → Incident → Lab)
- 통합 헌법(CONSTITUTION)으로 에이전트 간 프로토콜 표준화
- 일일 연대기(Chronicle)로 전 과정 기록 — 감사/추적 가능
- CEO 명령 체계 — 사람이 최종 결정권자
- BIP(Build in Public) 철학

**가재 컴퍼니의 강점:** 체계적이나 **13개는 1인 기업 초기에 과잉**. 단계적 확장이 필요.

### 1-2. OpenClaw 멀티 에이전트 구조

OpenClaw는 네이티브로 멀티 에이전트를 지원:
- `openclaw agents add <name>` — 에이전트 추가
- 각 에이전트: 독립된 workspace, SOUL.md, 세션, 인증
- 바인딩으로 채널 라우팅 (텔레그램/슬랙 등)
- 서브에이전트 스폰 가능 (메인 → 서브)

**커뮤니티 추천 기본 구조:** orchestrator + coder + researcher + automator (4-agent)

### 1-3. CrewAI 등 멀티 에이전트 프레임워크

CrewAI의 역할 기반 협업 패턴:
- **역할(Role)** + **목표(Goal)** + **배경(Backstory)** 정의
- 에이전트 간 태스크 위임
- Flow(전체 흐름) 안에 Crew(팀) 배치

---

## 2. SniffOps 에이전트 팀 설계

### 설계 원칙

1. **zzuckerfrei = CEO** — 아이디어, 아키텍처, 의사결정, 코드 리뷰
2. **단계적 확장** — 시작은 최소, 필요할 때 추가
3. **모델 비용 최적화** — 핵심 작업만 Opus, 나머지는 Sonnet/Haiku
4. **명확한 책임 분리** — 각 에이전트는 하나의 역할에 집중
5. **문서 기반 소통** — 에이전트 간 파일 시스템으로 비동기 협업

---

## 3. Phase 1: 최소 팀 (4 에이전트) — 즉시 시작

> 목표: SniffOps MVP 개발 + 기본 리서치

### 에이전트 구성

| ID | 이름 | 역할 | 모델 | 설명 |
|----|------|------|------|------|
| `main` | **🐕 Sniff** | CEO 비서 / 오케스트레이터 | Claude Opus | zzuckerfrei의 메인 에이전트. 작업 분배, 일정 관리, 의사결정 지원. 텔레그램 채널 담당 |
| `backend` | **🔧 Gopher** | 백엔드 개발자 | Claude Sonnet | Go 전문. MCP 서버, K8s client-go, SQLite, API 개발 |
| `frontend` | **🎨 Pixel** | 프론트엔드 개발자 | Claude Sonnet | React/Vite 전문. 웹 UI, 대시보드, 타임라인 뷰 |
| `researcher` | **🔍 Scout** | 리서처 | Claude Sonnet | 기술 조사, 경쟁사 분석, 트렌드 모니터링, 무엇이든 조사 |

### 상호작용 방식

```
zzuckerfrei (텔레그램) ←→ Sniff (main)
                     ↓ 서브에이전트 스폰
              ┌──────┼──────┐
              ↓      ↓      ↓
          Gopher   Pixel   Scout
         (backend) (frontend) (researcher)
```

- **Sniff**: zzuckerfrei의 요청을 해석 → 적절한 에이전트에 서브에이전트로 위임
- **Gopher**: `projects/sniffops/` 디렉토리에서 Go 코드 작업
- **Pixel**: `projects/sniffops/web/` 디렉토리에서 React 코드 작업
- **Scout**: 조사 결과를 마크다운 문서로 정리

### 각 에이전트 SOUL.md 핵심

#### Sniff (main) — 이미 운영 중
```markdown
# Sniff — CEO 비서 & 오케스트레이터
- zzuckerfrei의 의도를 정확히 파악하고 실행
- 복잡한 작업은 서브에이전트에 위임
- 매일 아침 8시 브리핑
- 프로젝트 진행상황 추적
```

#### Gopher (backend)
```markdown
# Gopher — Go 백엔드 개발자
- 전문: Go, MCP SDK, K8s client-go, SQLite
- 코드 스타일: 표준 Go 컨벤션, 에러 핸들링 철저
- 작업 완료 시 PR 요약 + 테스트 결과 보고
- 참고: projects/sniffops/RESEARCH.md의 아키텍처 결정 준수
```

#### Pixel (frontend)
```markdown
# Pixel — React 프론트엔드 개발자
- 전문: React, TypeScript, Vite, TailwindCSS
- 디자인: 깔끔하고 기능적인 대시보드 UI
- 컴포넌트 기반 설계, 재사용성 중시
- API 연동 시 Gopher가 만든 API 스펙 참조
```

#### Scout (researcher)
```markdown
# Scout — 만능 리서처
- 웹 검색 + 문서 분석으로 깊이 있는 조사
- 결과는 항상 마크다운 문서로 정리
- 경쟁사/시장 동향/기술 트렌드 모니터링
- Brave Search 무료 플랜 (초당 1건 제한) 주의
```

### 셋업 명령어

```bash
# 백엔드 에이전트
openclaw agents add backend --workspace ~/.openclaw/workspace

# 프론트엔드 에이전트
openclaw agents add frontend --workspace ~/.openclaw/workspace

# 리서처 에이전트
openclaw agents add researcher --workspace ~/.openclaw/workspace
```

> **참고:** 같은 workspace를 공유하면 프로젝트 파일에 모두 접근 가능. 에이전트별 독립 실행은 서브에이전트 스폰 방식으로 처리.

---

## 4. Phase 2: 확장 팀 (7 에이전트) — MVP 완성 후

> 목표: 품질 보증 + 문서화 + 출시 준비
> 시점: v0.1 베타 이후

### 추가 에이전트

| ID | 이름 | 역할 | 모델 | 설명 |
|----|------|------|------|------|
| `qa` | **🧪 Tester** | QA 엔지니어 | Claude Sonnet | 테스트 작성, 버그 리포트, 코드 리뷰 보조 |
| `writer` | **📝 Scribe** | 테크니컬 라이터 | Claude Sonnet | README, 문서, 블로그 포스트, CHANGELOG |
| `devops` | **🚀 Deploy** | DevOps 엔지니어 | Claude Sonnet | CI/CD, GitHub Actions, Docker, 릴리스 자동화 |

### 업데이트된 구조

```
zzuckerfrei (CEO) ←→ Sniff (오케스트레이터)
                    ↓
    ┌───────┬───────┼───────┬───────┬───────┐
    ↓       ↓       ↓       ↓       ↓       ↓
 Gopher  Pixel  Tester  Scribe  Deploy  Scout
 (백엔드) (프론트) (QA)   (문서)  (배포)  (리서치)
```

### 워크플로우 예시: 새 기능 개발

```
1. zzuckerfrei → Sniff: "위험도 평가 기능 추가해줘"
2. Sniff → Scout: 관련 기술 조사 (서브에이전트)
3. Sniff → Gopher: 백엔드 구현 (서브에이전트)
4. Sniff → Pixel: UI 구현 (서브에이전트)
5. Gopher/Pixel 완료 → Sniff → Tester: 테스트 (서브에이전트)
6. Tester 완료 → Sniff → Scribe: 문서 업데이트
7. zzuckerfrei 코드 리뷰 → Sniff → Deploy: 릴리스
```

---

## 5. Phase 3: 풀 컴퍼니 (10+ 에이전트) — 서비스 런칭 후

> 목표: 1인 기업 본격 운영
> 시점: v0.3+ / 첫 유료 고객 확보 시

### 추가 에이전트

| ID | 이름 | 역할 | 모델 | 설명 |
|----|------|------|------|------|
| `marketer` | **📢 Herald** | 마케터 | Claude Sonnet | 콘텐츠 마케팅, SNS, Product Hunt 런칭, SEO |
| `support` | **💬 Helper** | 고객 지원 | Claude Haiku | GitHub Issues 답변, FAQ 관리, 커뮤니티 |
| `analyst` | **📊 Metrics** | 비즈니스 분석가 | Claude Sonnet | 사용자 지표, 매출 분석, 시장 데이터 |

### 선택적 에이전트 (필요 시)

| ID | 이름 | 역할 | 모델 | 설명 |
|----|------|------|------|------|
| `legal` | **⚖️ Counsel** | 법무 | Claude Sonnet | 라이선스, 약관, 개인정보처리방침 |
| `designer` | **🎨 Canvas** | 디자이너 | Claude Sonnet | 로고, 랜딩페이지 디자인, 브랜딩 |
| `security` | **🛡️ Guard** | 보안 엔지니어 | Claude Sonnet | 코드 보안 리뷰, 취약점 분석 |

### 풀 컴퍼니 조직도

```
┌─────────────────────────────────────────────┐
│              zzuckerfrei (Human CEO)                │
│    아이디어 · 아키텍처 · 의사결정 · 코드리뷰   │
└──────────────────┬──────────────────────────┘
                   │
         ┌─────────┴─────────┐
         ↓                   ↓
    🐕 Sniff              📊 Metrics
   (COO/오케스트레이터)    (비즈니스 분석)
         │
    ┌────┼────┬────────┬──────────┐
    ↓    ↓    ↓        ↓          ↓
  개발팀  품질팀   비즈니스팀   운영팀
    │     │       │            │
  Gopher Tester  Herald      Helper
  Pixel  Scribe  Scout       Guard
  Deploy         Counsel     Canvas
```

---

## 6. 모델 전략 & 비용 최적화

### 모델 배분

| 모델 | 용도 | 에이전트 | 월 예상 |
|------|------|---------|---------|
| **Claude Opus** | 복잡한 의사결정, 아키텍처 | Sniff (main) | Max 플랜 포함 |
| **Claude Sonnet** | 코딩, 분석, 조사 | Gopher, Pixel, Scout 등 | Max 플랜 포함 |
| **Claude Haiku** | 단순 응답, 분류 | Helper (CS) | 최소 비용 |

### 비용 절감 팁
- **서브에이전트 방식**: 상시 구동이 아니라 필요 시 스폰 → 비용 절감
- **Max 플랜 활용**: 월정액에 포함된 사용량 최대한 활용
- **Haiku 활용**: 단순 작업은 Haiku로 처리
- **캐싱**: 반복 조사 결과는 마크다운 파일로 저장

---

## 7. 파일 시스템 기반 협업 프로토콜

### 디렉토리 구조

```
~/.openclaw/workspace/
├── projects/
│   └── sniffops/
│       ├── README.md              # 기획 문서
│       ├── RESEARCH.md            # 기술 조사
│       ├── AGENT-TEAM.md          # 이 문서
│       ├── CHANGELOG.md           # 변경 이력
│       ├── tasks/                 # 태스크 관리
│       │   ├── backlog.md         # 백로그
│       │   ├── in-progress.md     # 진행 중
│       │   └── done.md            # 완료
│       ├── docs/                  # 제품 문서
│       └── reports/               # 분석 리포트
├── memory/                        # 일일 기록
└── research/                      # 범용 리서치 결과
```

### 에이전트 간 소통 규칙

1. **작업 요청**: Sniff가 서브에이전트 스폰 시 명확한 태스크 정의
2. **결과 보고**: 서브에이전트는 파일로 결과 저장 + 텔레그램 알림
3. **코드 리뷰**: Gopher/Pixel 코드 → Tester 리뷰 → zzuckerfrei 최종 확인
4. **문서화**: 모든 결정은 마크다운으로 기록 (가재 컴퍼니의 Chronicle 참고)

---

## 8. 가재 컴퍼니 vs SniffOps 팀 비교

| 항목 | 가재 컴퍼니 | SniffOps 팀 |
|------|-----------|------------|
| 에이전트 수 | 13 (처음부터 풀) | 4 → 7 → 10+ (단계적) |
| 거버넌스 | 헌법 + 감사 체계 | 경량 프로토콜 |
| 기록 | 일일 연대기 (모든 대화 박제) | 일일 메모리 + 태스크 보드 |
| 복잡도 | 높음 (학습 곡선 큼) | 낮음 → 점진적 증가 |
| 적합한 상황 | 팀 운영 쇼케이스, BIP | 실제 제품 개발 + 1인 기업 |

**SniffOps 팀의 차별점:**
- 가재 컴퍼니의 체계성(계층 구조, 기록 문화)을 차용하되 **경량화**
- Phase 1은 4명으로 시작 → 실제로 필요해질 때만 확장
- 서브에이전트 방식으로 비용 효율적 운영

---

## 9. 즉시 실행 계획

### 오늘 할 일
1. ✅ 에이전트 팀 구성 제안서 작성 (이 문서)
2. Phase 1 에이전트 SOUL.md 파일 작성
3. `openclaw agents add` 로 에이전트 등록

### 이번 주
4. Gopher에게 Go 프로젝트 초기화 위임
5. Scout에게 MCP Go SDK 심층 분석 위임
6. 첫 번째 Tool (`sniff_get`) 구현 시작

### 이번 달
7. MVP v0.1 기능 구현 완료
8. Phase 2 에이전트 추가 (QA, 문서, DevOps)
9. GitHub 공개 + README 정비

---

## 참고 자료

- [가재 컴퍼니 (yuna-openclaw)](https://github.com/yuna-studio/yuna-openclaw) — 13-agent 풀 컴퍼니 사례
- [OpenClaw Multi-Agent Routing](https://docs.openclaw.ai/concepts/multi-agent) — 공식 멀티에이전트 문서
- [CrewAI](https://www.crewai.com/) — 역할 기반 멀티에이전트 프레임워크
- [OpenClaw Agent Team Guide](https://ai2sql.io/how-to-build-your-own-ai-agent-team-with-openclaw-in-15-minutes) — 커뮤니티 가이드

---

## 10. 추가 사례 분석

> 2026-02-11 보강: 주요 멀티에이전트 프레임워크 및 실제 사례 심층 조사

### 10-1. MetaGPT — "AI 소프트웨어 회사"

**출처:** [FoundationAgents/MetaGPT](https://github.com/FoundationAgents/MetaGPT) (arXiv 2308.00352)

**에이전트 구성 (5개 역할):**

| 역할 | 기능 |
|------|------|
| **Product Manager** | 사용자 요구사항 → 사용자 스토리, 경쟁 분석 |
| **Architect** | 시스템 설계, 데이터 구조, API 인터페이스 정의 |
| **Project Manager** | 태스크 분해, 일정 관리 |
| **Engineer** | 실제 코드 구현 |
| **QA Engineer** | 테스트 코드 작성, 버그 탐지 |

**상호작용 방식:**
- **SOP(표준운영절차) 기반 어셈블리 라인**: 각 역할이 순차적으로 아티팩트를 생성하며 다음 역할로 전달
- 한 줄 요구사항 입력 → PRD, 설계 문서, 태스크 목록, 코드 리포 자동 생성
- "구조화된 출력(Structured Output)" 강제로 할루시네이션 감소

**장점:**
- 실제 소프트웨어 회사의 프로세스를 그대로 시뮬레이션
- SOP로 에이전트 간 소통 품질 보장
- 2025년 MGX(MetaGPT X)로 진화 — 실제 서비스 제공

**단점:**
- 순차적 파이프라인이라 병렬 작업에 약함
- 요구사항이 모호하면 초기 PM 단계에서 품질 저하
- 자체 프레임워크 종속

**SniffOps 인사이트:**
- **SOP 기반 아티팩트 전달 패턴** 도입 가치 높음 — Gopher/Pixel 간 API 스펙을 문서로 명시적 전달
- 5개 역할이 SniffOps Phase 1-2와 거의 일치 → 검증된 구성

---

### 10-2. ChatDev — "가상 소프트웨어 회사"

**출처:** [ChatDev.ai](https://chatdev.ai/) / OpenBMB 연구팀

**에이전트 구성:**

| 역할 | 기능 |
|------|------|
| **CEO** | 전체 방향 설정, 태스크 정의 |
| **CTO** | 기술 결정, 아키텍처 |
| **Programmer** | 코드 구현 |
| **Art Designer** | UI/그래픽 에셋 |
| **Tester** | 테스트 실행, 버그 리포트 |

**상호작용 방식:**
- **Chat Chain**: 페어(2인) 대화 체인으로 각 개발 단계 진행
- 4단계 워터폴: Designing → Coding → Testing → Documenting
- 역할 간 자연어 대화로 의사결정

**장점:**
- 페어 대화가 직관적이고 디버깅 용이
- 가벼운 구현 (Python + ChatGPT API)
- 오픈소스로 커스터마이징 자유

**단점:**
- 단순한 앱 생성에 특화 (복잡한 프로젝트 한계)
- 대화 체인이 길어지면 컨텍스트 손실
- 실제 프로덕션 코드 품질은 미흡

**SniffOps 인사이트:**
- **페어 대화 패턴** — 복잡한 결정 시 두 에이전트가 토론하는 방식 적용 가능 (예: Gopher + Tester 페어 리뷰)
- CEO/CTO 분리 모델은 SniffOps에서 zzuckerfrei(CEO) + Sniff(CTO급) 구조와 유사

---

### 10-3. Microsoft AutoGen — "대화 패턴 프레임워크"

**출처:** [microsoft/autogen](https://github.com/microsoft/autogen)

**에이전트 구성 (유연한 패턴):**

| 패턴 | 구조 |
|------|------|
| **Two-Agent Chat** | 1:1 대화 (가장 단순) |
| **Sequential Chat** | A→B→C 순차 체인 |
| **Group Chat** | N명이 자유 토론 |
| **SelectorGroupChat** | LLM이 다음 발화자를 동적 선택 |
| **Nested Chat** | 에이전트 안에 에이전트 팀 내장 |

**상호작용 방식:**
- **ConversableAgent** 기반 — 모든 에이전트가 대화 가능
- UserProxyAgent(사람 대리) + AssistantAgent(AI) 조합
- GroupChatManager가 발화 순서 관리
- 0.4 버전에서 SelectorGroupChat 도입 — LLM이 맥락에 따라 최적 에이전트 선택

**장점:**
- 매우 유연한 대화 패턴 조합
- Microsoft 생태계 통합 (Azure, O365)
- 코드 실행 환경 내장

**단점:**
- 유연성이 높은 만큼 설계 복잡도 증가
- Group Chat에서 토큰 소비 폭증 가능
- 0.2 → 0.4 전환기로 API 불안정

**SniffOps 인사이트:**
- **SelectorGroupChat 패턴** — Sniff가 맥락에 따라 에이전트를 동적 선택하는 것과 유사. OpenClaw 서브에이전트 스폰에 이 개념 적용 가능
- **Nested Chat** — 서브에이전트 안에서 또 서브에이전트를 스폰하는 구조. 복잡한 태스크 분해에 유용

---

### 10-4. OpenAI Swarm — "경량 핸드오프 프레임워크"

**출처:** [openai/swarm](https://github.com/openai/swarm) (교육용/실험용)

**에이전트 구성:**
- 고정 역할 없음 — **Agent + Handoff Function** 조합으로 자유 구성
- 각 Agent는 instructions(시스템 프롬프트) + functions(도구) 보유
- Handoff: 에이전트 A → 에이전트 B로 제어권 이전

**상호작용 방식:**
- **Stateless**: 매 요청마다 독립 (상태 유지 안 함)
- **Explicit Handoff**: 함수 호출로 명시적 에이전트 전환
- 오케스트레이터 없이 에이전트끼리 직접 패스

**장점:**
- 극도로 가벼움 (~100줄 핵심 코드)
- 핸드오프가 투명하고 디버깅 쉬움
- 에이전트 추가/제거가 자유로움

**단점:**
- 프로덕션용 아님 (교육/프로토타입용)
- 상태 관리, 영속성 없음
- 복잡한 워크플로우 표현 한계

**SniffOps 인사이트:**
- **Handoff 패턴** 자체는 매우 유용 — OpenClaw 서브에이전트 스폰이 사실상 핸드오프
- **Stateless 원칙** — 각 서브에이전트가 독립적으로 작업하고 결과만 파일로 남기는 SniffOps 방식과 잘 맞음
- 가볍게 시작하고 필요 시 확장하는 철학이 SniffOps Phase 1과 동일

---

### 10-5. LangGraph — "그래프 기반 멀티에이전트 워크플로우"

**출처:** [LangChain/LangGraph](https://blog.langchain.com/langgraph-multi-agent-workflows/)

**에이전트 구성 패턴:**

| 패턴 | 설명 |
|------|------|
| **Supervisor** | 상위 에이전트가 하위 에이전트에 태스크 위임 |
| **Hierarchical** | 다단계 슈퍼바이저 (슈퍼바이저의 슈퍼바이저) |
| **Peer-to-Peer** | 에이전트끼리 직접 메시지 교환 |
| **Swarm** | 핸드오프 기반 에이전트 간 제어 이동 |

**상호작용 방식:**
- **State Graph**: 노드(에이전트) + 엣지(전이 조건)로 워크플로우 정의
- Supervisor를 "에이전트의 도구가 다른 에이전트"로 구현
- 조건부 분기, 루프, 병렬 실행 지원
- langgraph-supervisor 패키지로 간편 구성

**장점:**
- 복잡한 워크플로우를 시각적으로 표현 가능
- 상태 관리 내장 (체크포인트, 롤백)
- LangChain 생태계 전체 활용

**단점:**
- 학습 곡선 높음
- Python 종속
- 그래프가 복잡해지면 디버깅 어려움

**SniffOps 인사이트:**
- **Supervisor 패턴** = Sniff 오케스트레이터 모델과 동일
- **Hierarchical 패턴** — Phase 3에서 팀별 리드 에이전트를 두는 구조에 적용 가능
- 상태 그래프 개념을 파일 시스템 기반으로 구현하는 것이 SniffOps의 차별점

---

### 10-6. Google ADK (Agent Development Kit) — "계층적 에이전트 시스템"

**출처:** [Google ADK](https://google.github.io/adk-docs/) (2025.04 공개)

**에이전트 3대 카테고리:**

| 유형 | 역할 |
|------|------|
| **LLM Agent** | 두뇌 — 자연어 이해, 의사결정, 도구 호출 |
| **Workflow Agent** | 매니저 — Sequential, Parallel, Loop 오케스트레이션 |
| **Custom Agent** | 스페셜리스트 — 특수 로직 구현 |

**상호작용 방식:**
- **계층 트리**: 루트 에이전트 → sub_agents 리스트로 하위 에이전트 연결
- **LLM-Driven Delegation**: LLM이 sub_agent의 description을 보고 자동 라우팅
- **AgentTool**: 에이전트를 도구처럼 호출 (명시적)
- Vertex AI 통합으로 프로덕션 배포

**장점:**
- Google Cloud 네이티브 통합
- 3가지 에이전트 타입으로 깔끔한 분류
- Workflow Agent로 예측 가능한 파이프라인 구성

**단점:**
- Google Cloud 종속
- 아직 초기 (v0.5.0)
- LangGraph 대비 커뮤니티 작음

**SniffOps 인사이트:**
- **3가지 에이전트 타입 분류** 참고 — Sniff=LLM Agent, 워크플로우는 파일 기반으로 구현, Gopher/Pixel=Custom Agent
- **Description 기반 자동 라우팅** — 각 에이전트 SOUL.md에 명확한 설명을 넣으면 Sniff가 자연스럽게 위임 가능

---

### 10-7. Devin (Cognition AI) — "AI 소프트웨어 엔지니어"

**출처:** [Cognition AI / Devin](https://devin.ai/)

**에이전트 구성:**
- 단일 에이전트이나 내부적으로 **멀티에이전트 디스패치** 기능 보유
- 하나의 Devin이 하위 태스크를 다른 Devin 인스턴스에 위임
- 자체 신뢰도 평가 → 불확실하면 사람에게 질문

**상호작용 방식:**
- 슬랙/웹 인터페이스로 사람이 태스크 할당
- 내장 브라우저, 터미널, 에디터로 자율 작업
- Devin Wiki/Search로 코드베이스 이해 자동화

**장점:**
- end-to-end 자율 개발 (코딩→테스트→배포)
- 실제 기업 사례: Nubank에서 12x 효율 향상
- 장기 추론/계획 능력

**단점:**
- 유료 서비스 ($500/mo~)
- 블랙박스 (내부 동작 불투명)
- 복잡한 아키텍처 결정에는 여전히 사람 필요

**SniffOps 인사이트:**
- **자체 신뢰도 평가** 패턴 — 에이전트가 불확실할 때 zzuckerfrei에게 질문하는 프로토콜 도입 가치
- **멀티에이전트 디스패치** — Sniff가 서브에이전트를 스폰하는 것과 동일 패턴
- Devin은 "만능 단일 에이전트" → SniffOps는 "전문 분업 팀" — 차별화 포인트

---

### 10-8. General Intelligence — "1인 유니콘을 위한 AI 에이전트 팀"

**출처:** [Forbes, 2025.12](https://www.forbes.com/sites/stevenwolfepereira/2025/12/08/building-a-one-person-unicorn-this-startup-just-raised-87m-to-help/) / USV 투자

**에이전트 구성:**
- **Superoptimizer**: AI 에이전트들을 조율하는 오케스트레이션 시스템
- 제품 개발, 코드 리뷰, 고객 커뮤니케이션 등 비즈니스 기능별 에이전트 배치
- Sam Altman의 "노트북 하나로 10억 달러 기업" 비전 실현 목표

**상호작용 방식:**
- Superoptimizer가 전체 비즈니스 프로세스 조율
- 각 에이전트가 자율적으로 업무 수행, 사람은 감독/의사결정

**장점:**
- 1인 기업에 특화된 설계
- $8.7M 투자로 실제 검증 중
- 비즈니스 전체를 커버하는 야심찬 비전

**단점:**
- 아직 초기 (2025년 12월 라운드)
- 상세 아키텍처 비공개
- 완전 자율은 아직 먼 목표

**SniffOps 인사이트:**
- **SniffOps가 바로 이 비전의 실현체** — 1인 기업(zzuckerfrei)이 AI 에이전트 팀으로 제품 개발
- **Superoptimizer ≈ Sniff** — 동일한 오케스트레이터 개념
- "1인 유니콘" 트렌드가 시장에서 검증되고 있음 → SniffOps의 방향성 확인

---

### 10-9. CrewAI 실전 사례 모음

**출처:** [crewAI-examples](https://github.com/crewAIInc/crewAI-examples)

**대표 팀 구성 예시:**

| 유스케이스 | 에이전트 구성 | 상호작용 |
|-----------|-------------|---------|
| **Game Builder Crew** | Designer + Developer + Tester | 순차 파이프라인 |
| **Marketing Strategy** | Researcher + Strategist + Writer | 계층적 위임 |
| **Landing Page Generator** | Copywriter + Designer + Developer | Flow 기반 |
| **Instagram Post** | Content Strategist + Visual Creator + Copywriter | 병렬 → 합류 |
| **Stock Analysis** | Data Analyst + Trading Strategist + Risk Assessor | 순차 분석 체인 |

**CrewAI Flow + Crew 조합:**
- **Flow**: 전체 비즈니스 프로세스 (상태 관리)
- **Crew**: Flow 안의 각 단계를 처리하는 에이전트 팀
- Process 모드: `sequential` (순차) / `hierarchical` (계층적)

**장점:**
- Role + Goal + Backstory로 에이전트 성격 명확 정의
- 다양한 실전 예시 오픈소스
- Enterprise(AMP) 버전으로 프로덕션 지원

**단점:**
- Python 전용
- 복잡한 Flow는 디버깅 어려움
- LLM 비용 관리 필요

**SniffOps 인사이트:**
- **Role + Goal + Backstory** 패턴 → SOUL.md에 이미 적용 중
- **Game Builder Crew** 구성이 SniffOps Phase 1 (Designer + Developer + Tester)과 거의 동일
- Flow 개념 → tasks/ 디렉토리의 backlog → in-progress → done 파이프라인과 대응

---

### 사례 종합 비교표

| 프레임워크 | 에이전트 수 | 핵심 패턴 | 오케스트레이션 | SniffOps 유사도 |
|-----------|-----------|----------|-------------|---------------|
| **MetaGPT** | 5 | SOP 어셈블리 라인 | 순차 파이프라인 | ⭐⭐⭐⭐ |
| **ChatDev** | 5 | 페어 대화 체인 | 워터폴 | ⭐⭐⭐ |
| **AutoGen** | 유동적 | 그룹 채팅 / 선택적 | LLM 동적 선택 | ⭐⭐⭐ |
| **OpenAI Swarm** | 유동적 | 핸드오프 | 분산 (오케스트레이터 없음) | ⭐⭐ |
| **LangGraph** | 유동적 | 상태 그래프 | Supervisor / 계층적 | ⭐⭐⭐⭐ |
| **Google ADK** | 유동적 | 계층 트리 | Description 기반 라우팅 | ⭐⭐⭐ |
| **Devin** | 1(+내부 멀티) | 자율 에이전트 | 자체 디스패치 | ⭐⭐ |
| **General Intelligence** | 다수 | Superoptimizer | 중앙 오케스트레이터 | ⭐⭐⭐⭐⭐ |
| **CrewAI** | 유동적 | Role/Goal/Backstory | Flow + Crew | ⭐⭐⭐⭐ |
| **가재 컴퍼니** | 13 | 7계층 거버넌스 | 헌법 + CEO 명령 | ⭐⭐⭐⭐ |

### SniffOps에 적용할 핵심 인사이트 요약

1. **SOP 기반 아티팩트 전달** (MetaGPT) — 에이전트 간 명시적 문서 전달 프로토콜
2. **페어 대화 리뷰** (ChatDev) — 중요 결정 시 2개 에이전트 토론
3. **동적 에이전트 선택** (AutoGen SelectorGroupChat) — Sniff가 맥락에 따라 최적 에이전트 스폰
4. **Stateless 핸드오프** (Swarm) — 서브에이전트는 독립적, 결과만 파일로 전달
5. **Supervisor 패턴** (LangGraph) — Sniff 오케스트레이터의 이론적 근거
6. **신뢰도 기반 에스컬레이션** (Devin) — 불확실하면 사람에게 질문
7. **SOUL.md = Role + Goal + Backstory** (CrewAI) — 이미 적용 중, 더 정교화
8. **단계적 확장** (공통) — 4 → 7 → 10+ 전략이 업계 베스트 프랙티스와 일치
9. **1인 유니콘 트렌드** (General Intelligence) — SniffOps의 방향성이 시장에서 검증됨

---

_이 문서는 SniffOps 프로젝트의 에이전트 팀 구성 제안서입니다._
_프로젝트 진행에 따라 업데이트됩니다._
_최초 작성: 2026-02-11 | 사례 보강: 2026-02-11_
