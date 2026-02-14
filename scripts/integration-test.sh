#!/bin/bash
# SniffOps 통합 테스트 스크립트 (MCP JSON-RPC 방식)
# k3s 설치 + 테스트 리소스 생성 후 실행
# 사용법: bash scripts/integration-test.sh

set -e

SNIFFOPS_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SNIFFOPS_BIN="$SNIFFOPS_DIR/sniffops"
NAMESPACE="sniffops-test"
PASSED=0
FAILED=0
TOTAL=0

# Go/K8s 환경
export PATH="/home/smlee/go/bin:/home/smlee/gopath/bin:$PATH"
export GOROOT=/home/smlee/go
export GOPATH=/home/smlee/gopath
export KUBECONFIG=/home/smlee/.kube/config

# 색상
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# MCP JSON-RPC 테스트 함수
run_mcp_test() {
    local name="$1"
    local tool="$2"
    local args="$3"
    local expect="$4"
    local wait="${5:-3}"
    
    TOTAL=$((TOTAL + 1))
    echo -n "  [$TOTAL] $name ... "
    
    local response=$(
        (
            echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
            sleep 1
            echo "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}"
            sleep $wait
        ) | timeout $((wait + 5)) "$SNIFFOPS_BIN" serve 2>/dev/null | grep '"id":2'
    )
    
    if echo "$response" | grep -q "$expect"; then
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAIL${NC}"
        echo "    응답: $(echo "$response" | head -c 200)"
        echo "    기대: $expect"
        FAILED=$((FAILED + 1))
    fi
}

echo "🐝 SniffOps 통합 테스트 (MCP JSON-RPC)"
echo "========================================"
echo ""

# 바이너리 확인/빌드
if [ ! -f "$SNIFFOPS_BIN" ]; then
    echo "📦 바이너리 빌드 중..."
    cd "$SNIFFOPS_DIR"
    go build -o sniffops ./cmd/sniffops
    echo "  ✅ 빌드 완료"
    echo ""
fi

# kubeconfig 확인
if [ ! -f "$HOME/.kube/config" ]; then
    echo -e "${RED}❌ kubeconfig 없음. setup-k3s.sh 먼저 실행${NC}"
    exit 1
fi

# 파드 이름
POD_NAME=$(kubectl -n $NAMESPACE get pods -l app=nginx-test -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

echo "🧪 테스트 시작 (namespace: $NAMESPACE)"
echo ""

# --- Test: sniff_ping ---
echo -e "${YELLOW}[sniff_ping] 기본 연결 테스트${NC}"
run_mcp_test "ping 응답 확인" "sniff_ping" '{}' 'SniffOps MCP Server is running' 2

# --- Test: sniff_get ---
echo ""
echo -e "${YELLOW}[sniff_get] 리소스 조회 테스트${NC}"
run_mcp_test "Deployment 조회" "sniff_get" \
    "{\"kind\":\"Deployment\",\"namespace\":\"$NAMESPACE\",\"name\":\"nginx-test\"}" \
    'nginx-test' 5

run_mcp_test "Pod 목록 조회" "sniff_get" \
    "{\"kind\":\"Pod\",\"namespace\":\"$NAMESPACE\"}" \
    'nginx-test' 5

run_mcp_test "ConfigMap 조회" "sniff_get" \
    "{\"kind\":\"ConfigMap\",\"namespace\":\"$NAMESPACE\",\"name\":\"test-config\"}" \
    'test-config' 5

run_mcp_test "Service 조회" "sniff_get" \
    "{\"kind\":\"Service\",\"namespace\":\"$NAMESPACE\",\"name\":\"nginx-test-svc\"}" \
    'nginx-test-svc' 5

# --- Test: sniff_logs ---
echo ""
echo -e "${YELLOW}[sniff_logs] 로그 조회 테스트${NC}"
if [ -n "$POD_NAME" ]; then
    run_mcp_test "Pod 로그 조회" "sniff_logs" \
        "{\"pod\":\"$POD_NAME\",\"namespace\":\"$NAMESPACE\",\"lines\":3}" \
        'logs' 5
else
    echo "  ⏭️  파드 아직 준비 안 됨, 스킵"
fi

# --- Test: sniff_scale ---
echo ""
echo -e "${YELLOW}[sniff_scale] 스케일 테스트${NC}"
run_mcp_test "replicas 3으로 스케일" "sniff_scale" \
    "{\"name\":\"nginx-test\",\"namespace\":\"$NAMESPACE\",\"replicas\":3}" \
    'scaled' 5

# 복구
run_mcp_test "replicas 2로 복구" "sniff_scale" \
    "{\"name\":\"nginx-test\",\"namespace\":\"$NAMESPACE\",\"replicas\":2}" \
    'scaled' 5

# --- Test: sniff_exec ---
echo ""
echo -e "${YELLOW}[sniff_exec] 명령 실행 테스트${NC}"
if [ -n "$POD_NAME" ]; then
    run_mcp_test "hostname 확인" "sniff_exec" \
        "{\"pod\":\"$POD_NAME\",\"namespace\":\"$NAMESPACE\",\"command\":[\"hostname\"]}" \
        'nginx-test' 5
else
    echo "  ⏭️  파드 아직 준비 안 됨, 스킵"
fi

# --- Test: sniff_traces ---
echo ""
echo -e "${YELLOW}[sniff_traces] 트레이스 조회 테스트${NC}"
run_mcp_test "트레이스 목록 조회" "sniff_traces" '{"limit":10}' 'traces' 3

# --- Test: sniff_stats ---
echo ""
echo -e "${YELLOW}[sniff_stats] 통계 조회 테스트${NC}"
run_mcp_test "사용 통계 조회" "sniff_stats" '{}' 'total_traces' 3

# --- Test: sniff_delete ---
echo ""
echo -e "${YELLOW}[sniff_delete] 삭제 테스트${NC}"
# configmap 재생성 (이전 테스트에서 삭제됐을 수 있음)
kubectl -n $NAMESPACE create configmap test-config --from-literal=key1=value1 2>/dev/null || true
run_mcp_test "ConfigMap 삭제" "sniff_delete" \
    "{\"kind\":\"ConfigMap\",\"namespace\":\"$NAMESPACE\",\"name\":\"test-config\"}" \
    'deleted' 5

# configmap 재생성 (cleanup)
kubectl -n $NAMESPACE create configmap test-config --from-literal=key1=value1 2>/dev/null || true

# --- 결과 ---
echo ""
echo "========================================"
echo -e "📊 결과: ${GREEN}${PASSED} PASS${NC} / ${RED}${FAILED} FAIL${NC} / ${TOTAL} TOTAL"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "🎉 ${GREEN}전체 통과! SniffOps MVP 정상 동작!${NC}"
    exit 0
else
    echo -e "⚠️  ${RED}${FAILED}개 실패. 위 로그 확인 필요${NC}"
    exit 1
fi
