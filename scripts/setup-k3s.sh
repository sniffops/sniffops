#!/bin/bash
# SniffOps k3s 테스트 환경 셋업 스크립트
# 사용법: sudo bash scripts/setup-k3s.sh

set -e

echo "🐝 SniffOps 테스트 환경 셋업 시작"
echo "=================================="

# 1. k3s 설치
echo ""
echo "📦 [1/4] k3s 설치 중..."
if command -v k3s &> /dev/null; then
    echo "  ✅ k3s 이미 설치됨: $(k3s --version)"
else
    curl -sfL https://get.k3s.io | sh -
    echo "  ✅ k3s 설치 완료"
fi

# k3s 시작 대기
echo "  ⏳ k3s 시작 대기 중..."
sleep 10
until k3s kubectl get nodes &> /dev/null; do
    echo "  ... 아직 시작 중"
    sleep 5
done
echo "  ✅ k3s 정상 가동"

# 2. kubeconfig 설정 (smlee 유저용)
echo ""
echo "🔑 [2/4] kubeconfig 설정..."
SMLEE_HOME=$(eval echo ~smlee)
mkdir -p "$SMLEE_HOME/.kube"
cp /etc/rancher/k3s/k3s.yaml "$SMLEE_HOME/.kube/config"
chown smlee:smlee "$SMLEE_HOME/.kube/config"
chmod 600 "$SMLEE_HOME/.kube/config"
echo "  ✅ kubeconfig → $SMLEE_HOME/.kube/config"

# 3. 테스트용 네임스페이스 + 리소스 생성
echo ""
echo "🧪 [3/4] 테스트 리소스 생성 중..."

# 네임스페이스
k3s kubectl create namespace sniffops-test 2>/dev/null || true
echo "  ✅ namespace: sniffops-test"

# nginx 디플로이먼트 (테스트용)
k3s kubectl -n sniffops-test apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-test
  labels:
    app: nginx-test
    purpose: sniffops-testing
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx-test
  template:
    metadata:
      labels:
        app: nginx-test
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "32Mi"
            cpu: "50m"
          limits:
            memory: "64Mi"
            cpu: "100m"
EOF
echo "  ✅ deployment: nginx-test (replicas: 2)"

# 서비스
k3s kubectl -n sniffops-test apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: nginx-test-svc
spec:
  selector:
    app: nginx-test
  ports:
  - port: 80
    targetPort: 80
EOF
echo "  ✅ service: nginx-test-svc"

# configmap (삭제 테스트용)
k3s kubectl -n sniffops-test apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key1: value1
  key2: value2
EOF
echo "  ✅ configmap: test-config"

# 4. 상태 확인
echo ""
echo "📋 [4/4] 환경 확인..."
echo ""
echo "--- 노드 ---"
k3s kubectl get nodes
echo ""
echo "--- sniffops-test 네임스페이스 ---"
k3s kubectl -n sniffops-test get all
echo ""

echo "=================================="
echo "🎉 셋업 완료!"
echo ""
echo "다음 단계:"
echo "  1. SniffOps 빌드: cd projects/sniffops && go build -o sniffops ./cmd/sniffops"
echo "  2. 통합 테스트:   bash scripts/integration-test.sh"
echo ""
