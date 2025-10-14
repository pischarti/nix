# Testing Leader Election

This guide walks through testing the leader election functionality of the kaws operator.

## Prerequisites

- Kubernetes cluster (minikube, kind, or EKS)
- kubectl configured
- Docker (for building container images)

## Setup

### 1. Build and Push Image

```bash
# Build the operator image
make docker-build IMG=your-registry/kaws-operator:latest

# Push to registry
make docker-push IMG=your-registry/kaws-operator:latest
```

### 2. Deploy Operator

```bash
# Install CRD
make install

# Deploy operator with 3 replicas and leader election
make deploy

# Verify deployment
kubectl get deploy -n kube-system kaws-operator
kubectl get pods -n kube-system -l app=kaws-operator
```

Expected output:
```
NAME             READY   STATUS    RESTARTS   AGE
kaws-operator    3/3     Running   0          30s
```

## Test Scenarios

### Test 1: Leader Election on Startup

**Objective:** Verify that one replica is elected as leader on startup.

**Steps:**

1. Watch the lease resource:
```bash
kubectl get lease kaws-operator-lock -n kube-system -w
```

2. Check logs from all replicas:
```bash
kubectl logs -n kube-system -l app=kaws-operator --tail=50 | grep -i leader
```

**Expected Results:**
- One pod logs: `successfully acquired lease kube-system/kaws-operator-lock`
- Other pods log: `attempting to acquire leader lease` or `waiting to acquire leadership`
- Lease shows `holderIdentity` matching the leader pod name

**Verification:**
```bash
# Get the leader pod name
LEADER=$(kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.holderIdentity}')
echo "Current leader: $LEADER"

# Check that only leader is reconciling
kubectl logs -n kube-system $LEADER | grep "reconciling"
```

### Test 2: Leader Failover (Pod Deletion)

**Objective:** Verify automatic failover when leader pod is deleted.

**Steps:**

1. Identify the current leader:
```bash
LEADER=$(kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.holderIdentity}')
echo "Current leader: $LEADER"
```

2. Start watching the lease:
```bash
kubectl get lease kaws-operator-lock -n kube-system -w
```

3. In another terminal, delete the leader pod:
```bash
kubectl delete pod -n kube-system $LEADER
```

4. Observe the failover:
```bash
# Watch for new leader election
kubectl logs -n kube-system -l app=kaws-operator -f | grep -i leader
```

**Expected Results:**
- Old leader stops renewing lease
- After ~15 seconds, new leader is elected
- Lease `holderIdentity` updates to new pod name
- `leaseTransitions` increments by 1
- New leader starts reconciling

**Timing:**
- Lease expiration: ~15 seconds
- New leader acquisition: ~2 seconds
- Total failover time: ~17 seconds

### Test 3: Leader Failover (Node Failure)

**Objective:** Verify failover when the node running the leader fails.

**Steps:**

1. Identify leader and its node:
```bash
LEADER=$(kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.holderIdentity}')
NODE=$(kubectl get pod -n kube-system $LEADER -o jsonpath='{.spec.nodeName}')
echo "Leader: $LEADER on node: $NODE"
```

2. Cordon and drain the node:
```bash
kubectl cordon $NODE
kubectl drain $NODE --ignore-daemonsets --delete-emptydir-data
```

3. Watch for failover:
```bash
kubectl get lease kaws-operator-lock -n kube-system -w
```

**Expected Results:**
- Pod eviction triggers immediate failover
- New leader elected on different node
- Total failover: ~15-30 seconds

**Cleanup:**
```bash
kubectl uncordon $NODE
```

### Test 4: Rolling Update

**Objective:** Verify zero-downtime during rolling updates.

**Steps:**

1. Update the operator image:
```bash
kubectl set image deployment/kaws-operator -n kube-system operator=your-registry/kaws-operator:v2
```

2. Watch the rollout:
```bash
kubectl rollout status deployment/kaws-operator -n kube-system
```

3. Monitor leader changes:
```bash
kubectl get lease kaws-operator-lock -n kube-system -w
```

4. Check for continuous operation:
```bash
kubectl get eventrecycler -o yaml
# Check status.lastCheckTime - should update continuously
```

**Expected Results:**
- Rolling update completes successfully
- Leadership may transfer 1-3 times during update
- No extended downtime (max ~17 seconds per transfer)
- Status continues to update throughout

### Test 5: Split Brain Prevention

**Objective:** Verify only one replica reconciles at a time.

**Steps:**

1. Deploy a test EventRecycler:
```bash
kubectl apply -f config/samples/eventrecycler_sample.yaml
```

2. Enable debug logging on all pods:
```bash
for pod in $(kubectl get pods -n kube-system -l app=kaws-operator -o name); do
  kubectl logs -n kube-system $pod -f | grep -E "reconciling|leader" &
done
```

3. Watch for 5 minutes and verify:
```bash
# Count reconciliation events per pod
kubectl logs -n kube-system -l app=kaws-operator --tail=1000 | \
  grep -E "pod-[a-z0-9-]+.*reconciling" | \
  cut -d' ' -f1 | sort | uniq -c
```

**Expected Results:**
- Only one pod shows reconciliation logs
- Other pods show "waiting to acquire leadership"
- No duplicate reconciliations

### Test 6: Lease Transitions Under Load

**Objective:** Verify stable leadership under normal conditions.

**Steps:**

1. Let operator run for 1 hour
```bash
sleep 3600
```

2. Check lease transition count:
```bash
kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.leaseTransitions}'
```

**Expected Results:**
- Lease transitions: 0-2 (only initial election + any pod restarts)
- No frequent leader changes
- If transitions > 10: investigate instability

**Troubleshooting high transitions:**
```bash
# Check pod restarts
kubectl get pods -n kube-system -l app=kaws-operator

# Check resource usage
kubectl top pods -n kube-system -l app=kaws-operator

# Check events
kubectl get events -n kube-system --field-selector involvedObject.name=kaws-operator-lock
```

## Monitoring Commands

### Real-time Leader Status

```bash
# Watch lease
watch -n 1 'kubectl get lease kaws-operator-lock -n kube-system -o yaml | grep -A 5 spec'

# Watch operator pods
watch -n 1 'kubectl get pods -n kube-system -l app=kaws-operator'
```

### Leader Election Metrics

```bash
# Current leader
kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.holderIdentity}'

# Lease transitions (should be low)
kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.leaseTransitions}'

# Last renewal time
kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.renewTime}'

# Leader age
ACQUIRE_TIME=$(kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.acquireTime}')
echo "Leader acquired at: $ACQUIRE_TIME"
```

### Pod-specific Logs

```bash
# Get leader pod name
LEADER=$(kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.holderIdentity}')

# View leader logs
kubectl logs -n kube-system $LEADER -f

# View all follower logs
for pod in $(kubectl get pods -n kube-system -l app=kaws-operator -o name | grep -v $LEADER); do
  echo "=== Logs from $pod ==="
  kubectl logs -n kube-system $pod --tail=20
done
```

## Performance Testing

### Test Leader Election Overhead

**Measure API server load:**

```bash
# Before deploying operator
kubectl top nodes
kubectl get --raw /metrics | grep apiserver_request_total

# After deploying operator (with 3 replicas)
# Wait 5 minutes for metrics to accumulate
kubectl top nodes
kubectl get --raw /metrics | grep apiserver_request_total
```

**Expected Impact:**
- Minimal CPU increase (< 5%)
- 3-4 additional API requests per second (lease renewals)
- Memory increase: ~150-1500MB total across 3 replicas (for informer caches)

### Test Reconciliation Performance

```bash
# Create test EventRecycler
kubectl apply -f config/samples/eventrecycler_sample.yaml

# Measure reconciliation time
kubectl logs -n kube-system -l app=kaws-operator | grep -E "reconcile|duration"

# Check EventRecycler status update frequency
watch -n 1 'kubectl get eventrecycler -o yaml | grep -A 5 status'
```

**Expected Performance:**
- Reconciliation time: < 1 second (with cached informers)
- Status updates: Every 60 seconds (default watchInterval)
- No reconciliation on follower pods

## Common Issues

### Issue: No Leader Elected

**Symptoms:**
```bash
kubectl get lease kaws-operator-lock -n kube-system
# Returns: NotFound
```

**Diagnosis:**
```bash
# Check RBAC
kubectl auth can-i create leases \
  --as=system:serviceaccount:kube-system:kaws-operator \
  -n kube-system

# Check pod logs for errors
kubectl logs -n kube-system -l app=kaws-operator | grep -i error
```

**Fix:**
```bash
# Ensure leader election RBAC is installed
kubectl apply -f config/rbac/leader_election_role.yaml
```

### Issue: Frequent Leader Changes

**Symptoms:**
```bash
kubectl get lease kaws-operator-lock -n kube-system -o jsonpath='{.spec.leaseTransitions}'
# Returns: > 20 in first hour
```

**Diagnosis:**
```bash
# Check pod restarts
kubectl get pods -n kube-system -l app=kaws-operator

# Check resource limits
kubectl describe pods -n kube-system -l app=kaws-operator | grep -A 5 Limits

# Check node conditions
kubectl get nodes -o wide
```

**Fix:**
- Increase resource limits if pods are being OOMKilled
- Check network latency to API server
- Increase lease duration if network is unstable

### Issue: Multiple Replicas Reconciling

**Symptoms:**
- Duplicate events or reconciliations
- Multiple pods logging reconciliation activity

**Diagnosis:**
```bash
# Check if leader election is enabled
kubectl logs -n kube-system -l app=kaws-operator | grep "Leader election"

# Verify lease exists
kubectl get lease kaws-operator-lock -n kube-system
```

**Fix:**
- Ensure operator is started with `--use-crd` flag
- Verify leader election is enabled in code
- Check that all replicas are using the same LeaderElectionID

## Success Criteria

✅ **Basic Functionality:**
- [ ] One replica becomes leader within 30 seconds of deployment
- [ ] Lease shows correct leader identity
- [ ] Leader logs show reconciliation activity
- [ ] Followers log "waiting to acquire leadership"

✅ **Failover:**
- [ ] New leader elected within 17 seconds of leader deletion
- [ ] Lease transitions increment correctly
- [ ] No service disruption > 20 seconds

✅ **Stability:**
- [ ] Lease transitions < 5 in first hour
- [ ] No pod restarts or crashes
- [ ] Consistent reconciliation activity

✅ **High Availability:**
- [ ] Rolling updates complete with < 1 minute total downtime
- [ ] Node failures handled gracefully
- [ ] Can tolerate loss of 2 out of 3 replicas

## Cleanup

```bash
# Remove test resources
kubectl delete eventrecycler --all

# Undeploy operator
make undeploy

# Uninstall CRD
make uninstall
```

