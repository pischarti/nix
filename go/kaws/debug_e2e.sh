#!/bin/bash
# Debug E2E test failures

echo "Creating kind cluster..."
kind create cluster --name kaws-e2e-debug --wait 60s

echo ""
echo "Building and loading operator image..."
cd /Users/steve/dev/nix
docker build -f go/kaws/Dockerfile -t kaws-operator:test .
kind load docker-image kaws-operator:test --name kaws-e2e-debug

cd go/kaws

echo ""
echo "Installing CRD..."
kubectl apply -f config/crd/eventrecycler.yaml --context kind-kaws-e2e-debug

echo ""
echo "Installing RBAC..."
kubectl apply -f config/rbac/role.yaml --context kind-kaws-e2e-debug
kubectl apply -f config/rbac/leader_election_role.yaml --context kind-kaws-e2e-debug

echo ""
echo "Deploying operator..."
cat <<YAML | kubectl apply -f - --context kind-kaws-e2e-debug
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kaws-operator
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kaws-operator
  template:
    metadata:
      labels:
        app: kaws-operator
    spec:
      serviceAccountName: kaws-operator
      containers:
      - name: operator
        image: kaws-operator:test
        imagePullPolicy: Never
        command: ["/kaws"]
        args: ["operator", "--use-crd", "--verbose"]
        env:
        - name: AWS_REGION
          value: "us-east-1"
        - name: AWS_ACCESS_KEY_ID
          value: "test"
        - name: AWS_SECRET_ACCESS_KEY
          value: "test"
YAML

echo ""
echo "Waiting for pod..."
sleep 10

echo ""
echo "Pod status:"
kubectl get pods -n kube-system -l app=kaws-operator --context kind-kaws-e2e-debug

echo ""
echo "Pod describe:"
kubectl describe pods -n kube-system -l app=kaws-operator --context kind-kaws-e2e-debug | tail -50

echo ""
echo "Pod logs:"
kubectl logs -n kube-system -l app=kaws-operator --context kind-kaws-e2e-debug --tail=50

echo ""
echo "Cleanup:"
echo "  kind delete cluster --name kaws-e2e-debug"
