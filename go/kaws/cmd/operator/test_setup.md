# Operator Integration Tests

This directory contains integration tests for the kaws operator that use real Kubernetes (kind) and mock AWS services (localstack).

## Prerequisites

1. **Install kind (Kubernetes in Docker)**
   ```bash
   # macOS
   brew install kind
   
   # Linux
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind
   ```

2. **Install Docker**
   - Docker must be running for kind to work
   - https://docs.docker.com/get-docker/

3. **Install localstack (for AWS mocking)**
   ```bash
   # Using Docker (recommended)
   docker pull localstack/localstack
   
   # Or using pip
   pip install localstack
   ```

## Running the Tests

### 1. Start localstack

```bash
# Using Docker (recommended)
docker run -d \
  --name kaws-localstack \
  -p 4566:4566 \
  -e SERVICES=ec2,autoscaling \
  -e DEBUG=1 \
  localstack/localstack

# Verify it's running
curl http://localhost:4566/_localstack/health
```

### 2. Run the integration tests

```bash
# Navigate to the go/kaws directory
cd go/kaws

# Run integration tests using the Makefile (recommended)
make -f Makefile.integration test-integration

# Or run directly with go test
go test -v -tags=integration ./cmd/operator/

# Or run all tests including unit tests
go test -v ./cmd/operator/
```

### 3. Run specific tests

```bash
# Run only the error detection test
go test -v -tags=integration -run TestOperatorIntegration/DetectErrorEvents ./cmd/operator/

# Run with more detailed output
go test -v -tags=integration ./cmd/operator/ 2>&1 | tee test-output.log
```

## Test Structure

The integration tests verify:

1. **Error Event Detection** (`testDetectErrorEvents`)
   - Creates mock Kubernetes events
   - Verifies operator can detect them
   - Checks event processing logic

2. **Node Group Identification** (`testNodeGroupIdentification`)
   - Creates mock nodes with provider IDs
   - Creates mock EC2 instances in localstack
   - Verifies operator can map pods → nodes → instances → node groups

3. **Event Filtering** (`testEventFiltering`)
   - Tests various search terms
   - Verifies filtering logic works correctly
   - Ensures no false positives/negatives

## Test Environment

Each test run:

1. Creates a fresh kind cluster (`kaws-test`)
2. Connects to localstack for AWS APIs
3. Creates mock resources:
   - Kubernetes nodes
   - Kubernetes pods
   - Kubernetes events
   - (Optional) EC2 instances in localstack
4. Runs the test scenarios
5. Cleans up everything

## Cleanup

```bash
# Stop and remove localstack container
docker stop kaws-localstack
docker rm kaws-localstack

# Manually delete kind cluster if needed
kind delete cluster --name kaws-test
```

## Troubleshooting

### Kind cluster creation fails

```bash
# Check Docker is running
docker ps

# Check kind version
kind version

# Try deleting existing cluster
kind delete cluster --name kaws-test
```

### Localstack not accessible

```bash
# Check localstack is running
docker ps | grep localstack

# Check localstack logs
docker logs kaws-localstack

# Test connectivity
curl http://localhost:4566/_localstack/health
```

### Tests hang or timeout

```bash
# Increase test timeout
go test -v -tags=integration -timeout 10m ./cmd/operator/

# Check cluster is accessible
kubectl --context kind-kaws-test get nodes
```

## CI/CD Integration

For CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Setup kind
  run: |
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind

- name: Start localstack
  run: |
    docker run -d \
      --name localstack \
      -p 4566:4566 \
      -e SERVICES=ec2,autoscaling \
      localstack/localstack
    
    # Wait for localstack to be ready
    timeout 60 bash -c 'until curl -s http://localhost:4566/_localstack/health; do sleep 1; done'

- name: Run integration tests
  run: go test -v -tags=integration ./cmd/operator/
```

## Development Tips

1. **Skip integration tests during development**
   ```bash
   go test -short ./cmd/operator/
   ```

2. **Keep kind cluster running between tests**
   - Comment out the teardown code
   - Manually inspect resources

3. **Debug with kubectl**
   ```bash
   export KUBECONFIG=~/.kube/config
   kubectl config use-context kind-kaws-test
   kubectl get events --all-namespaces
   kubectl get pods -A
   ```

4. **Add more test scenarios**
   - Add new test functions to `operator_integration_test.go`
   - Follow the pattern of existing tests
   - Use descriptive test names

