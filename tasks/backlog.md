# 📋 Backlog

## [TASK-001] 아키텍처 초안 작성
- **담당**: Architect (cto)
- **우선순위**: P0
- **설명**: SniffOps 전체 시스템 아키텍처 설계. MCP 서버 구조, trace 저장 흐름, SQLite 스키마 초안.
- **산출물**: docs/architecture.md
- **참고**: RESEARCH.md

## [TASK-002] MCP Go SDK 심층 분석
- **담당**: Scout (researcher)
- **우선순위**: P0
- **설명**: MCP Go SDK (modelcontextprotocol/go-sdk) 최신 버전 심층 분석. Tool 등록, 핸들러 패턴, 에러 처리.
- **산출물**: research/mcp-go-sdk-deep-dive.md

## [TASK-003] Go 프로젝트 초기화
- **담당**: Gopher (backend)
- **우선순위**: P0
- **설명**: Go 모듈 초기화, 디렉토리 구조, 기본 main.go, Makefile.
- **산출물**: projects/sniffops/cmd/, go.mod, Makefile
- **선행**: TASK-001 (아키텍처 확정 후)

## [TASK-004] K8s MCP 서버 구현 패턴 조사
- **담당**: Scout (researcher)
- **우선순위**: P1
- **설명**: 기존 K8s MCP 서버들의 Tool 구현 패턴 심층 비교 (containers/kubernetes-mcp-server 중심)
- **산출물**: research/k8s-mcp-patterns.md
