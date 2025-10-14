# Leader Election in kaws Operator

## Overview

The `kaws` operator uses Kubernetes leader election to enable running multiple replicas for high availability. Only the elected leader actively reconciles resources, while follower replicas stand ready to take over if the leader fails.

## How Leader Election Works

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Kubernetes API Server                          │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Lease: kaws-operator-lock (kube-system namespace)       │   │
│  │  holder: kaws-operator-7d8f9c5b6b-abc12                  │   │
│  │  renewTime: 2024-01-15T10:30:45Z                         │   │
│  │  leaseDuration: 15s                                      │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                    ▲                    ▲                ▲
                    │                    │                │
                    │ Renew lease        │ Try acquire    │ Try acquire
                    │ every 2s           │ every 2s       │ every 2s
                    │                    │                │
        ┌───────────┴────────┐  ┌────────┴────────┐  ┌───┴────────────┐
        │  Replica 1         │  │  Replica 2      │  │  Replica 3     │
        │  (LEADER)          │  │  (FOLLOWER)     │  │  (FOLLOWER)    │
        │  ✓ Reconciling     │  │  ⏸ Watching     │  │  ⏸ Watching    │
        │  ✓ Cache synced    │  │  ✓ Cache synced │  │  ✓ Cache synced│
        └────────────────────┘  └─────────────────┘  └────────────────┘
```

### Election Process

1. **Startup**: Each replica starts and attempts to acquire the lease
2. **Leader elected**: First replica to acquire lease becomes leader
3. **Heartbeat**: Leader renews lease every 2 seconds
4. **Followers wait**: Other replicas check lease every 2 seconds, ready to take over
5. **Failover**: If leader stops renewing (crashed), lease expires after 15 seconds
6. **New leader**: First follower to acquire expired lease becomes new leader

### Configuration

The operator uses these leader election settings:

```go
ctrl.Options{
    LeaderElection:          true,
    LeaderElectionID:        "kaws-operator-lock",        // Lease name
    LeaderElectionNamespace: "kube-system",               // Lease namespace
    LeaseDuration:           15 * time.Second,            // How long lease is valid
    RenewDeadline:           10 * time.Second,            // Leader must renew within this
    RetryPeriod:             2 * time.Second,             // How often to check/renew
}
```

**Parameter explanations:**
- **LeaseDuration (15s)**: Maximum time a leader can be unresponsive before considered dead
- **RenewDeadline (10s)**: Leader must successfully renew within this time
- **RetryPeriod (2s)**: How frequently to attempt acquire/renew

### Lease Resource

Leader election uses a Kubernetes Lease resource in `kube-system`:

```yaml
apiVersion: coordination.k8s.io/v1
kind: Lease
metadata:
  name: kaws-operator-lock
  namespace: kube-system
spec:
  holderIdentity: kaws-operator-7d8f9c5b6b-abc12  # Current leader pod
  leaseDurationSeconds: 15
  acquireTime: "2024-01-15T10:30:30Z"
  renewTime: "2024-01-15T10:30:45Z"
  leaseTransitions: 2  # Number of times leadership has changed
```

## High Availability Setup

### Deployment Configuration

The operator is configured for 3 replicas by default:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kaws-operator
  namespace: kube-system
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: operator
        image: your-registry/kaws-operator:latest
        command:
        - /kaws
        args:
        - operator
        - --use-crd
```

### Required RBAC

Leader election requires permissions to manage leases:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kaws-operator-leader-election
  namespace: kube-system
rules:
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]  # For election events
```

## Monitoring Leader Election

### Check Current Leader

```bash
# View the lease to see who is leader
kubectl get lease kaws-operator-lock -n kube-system -o yaml

# Check which pod is leader from logs
kubectl logs -n kube-system -l app=kaws-operator | grep "leader"
```

### Leader Election Events

Watch for leader election events:

```bash
# View leader election events
kubectl get events -n kube-system --field-selector involvedObject.name=kaws-operator-lock

# Example events:
# kaws-operator-abc12 became leader
# kaws-operator-def34 started leading
```

### Replica Status

Check all replica statuses:

```bash
# List all operator pods
kubectl get pods -n kube-system -l app=kaws-operator

# Check logs from each pod
kubectl logs -n kube-system kaws-operator-7d8f9c5b6b-abc12
kubectl logs -n kube-system kaws-operator-7d8f9c5b6b-def34
kubectl logs -n kube-system kaws-operator-7d8f9c5b6b-ghi56
```

**Expected log output:**

**Leader pod:**
```
Starting manager with leader election
successfully acquired lease kube-system/kaws-operator-lock
✓ Leader election enabled - safe to run multiple replicas
ℹ️  Only the leader replica will reconcile resources
Starting controller manager with informers and leader election
```

**Follower pods:**
```
Starting manager with leader election
attempting to acquire leader lease kube-system/kaws-operator-lock
✓ Leader election enabled - safe to run multiple replicas
waiting to acquire leadership...
```

## Failure Scenarios

### Leader Pod Crashes

**Timeline:**
1. `T+0s`: Leader pod crashes
2. `T+0s - T+15s`: Leader fails to renew lease
3. `T+15s`: Lease expires
4. `T+15s - T+17s`: Follower acquires lease
5. `T+17s`: New leader starts reconciling

**Total downtime:** ~15-17 seconds

### Leader Node Fails

**Timeline:**
1. `T+0s`: Node running leader pod fails
2. `T+0s - T+40s`: Kubernetes detects node failure
3. `T+40s`: Pod marked as terminated
4. `T+40s`: Lease can no longer be renewed
5. `T+40s - T+55s`: Lease expires
6. `T+55s`: Follower acquires lease
7. `T+55s`: New leader starts reconciling

**Total downtime:** ~55 seconds (due to node failure detection delay)

### All Pods Restart Simultaneously

**Timeline:**
1. `T+0s`: All pods restart (e.g., rolling update)
2. `T+0s - T+30s`: Pods start up, informer caches sync
3. `T+30s`: First pod ready attempts to acquire lease
4. `T+30s`: New leader elected
5. `T+30s`: Leader starts reconciling

**Total downtime:** ~30 seconds (cache warm-up time)

## Tuning Leader Election

### For Faster Failover

Reduce lease duration for faster failover:

```go
LeaseDuration: ptr(8 * time.Second),   // Faster expiration
RenewDeadline: ptr(5 * time.Second),   // Tighter deadline
RetryPeriod:   ptr(1 * time.Second),   // More frequent checks
```

**Tradeoffs:**
- ✅ Faster failover (~8 seconds)
- ❌ More API server load (more frequent lease updates)
- ❌ Higher risk of false leader changes during network hiccups

### For Reduced API Load

Increase lease duration for less API load:

```go
LeaseDuration: ptr(30 * time.Second),  // Longer lease
RenewDeadline: ptr(20 * time.Second),  // More time to renew
RetryPeriod:   ptr(5 * time.Second),   // Less frequent checks
```

**Tradeoffs:**
- ✅ Less API server load
- ✅ More stable (fewer false leader changes)
- ❌ Slower failover (~30 seconds)

### Current Settings (Balanced)

The default configuration provides a good balance:

```go
LeaseDuration: ptr(15 * time.Second),  // ✓ Reasonable failover time
RenewDeadline: ptr(10 * time.Second),  // ✓ Enough time for retries
RetryPeriod:   ptr(2 * time.Second),   // ✓ Responsive without overload
```

## Best Practices

1. **Always run 3+ replicas** for high availability
2. **Spread across availability zones** using pod anti-affinity:
   ```yaml
   affinity:
     podAntiAffinity:
       preferredDuringSchedulingIgnoredDuringExecution:
       - weight: 100
         podAffinityTerm:
           topologyKey: topology.kubernetes.io/zone
           labelSelector:
             matchLabels:
               app: kaws-operator
   ```

3. **Monitor lease transitions** - frequent transitions indicate instability
4. **Set appropriate resource limits** to prevent OOM kills
5. **Use PodDisruptionBudget** to maintain availability during node drains:
   ```yaml
   apiVersion: policy/v1
   kind: PodDisruptionBudget
   metadata:
     name: kaws-operator
   spec:
     minAvailable: 1
     selector:
       matchLabels:
         app: kaws-operator
   ```

## Troubleshooting

### Problem: Frequent Leader Changes

**Symptoms:**
```bash
kubectl get lease kaws-operator-lock -n kube-system -o yaml
# leaseTransitions: 45  # Very high number
```

**Causes:**
- Network issues between pods and API server
- Resource limits too low (pods being throttled/OOMKilled)
- Node instability

**Solutions:**
- Increase lease duration
- Increase resource limits
- Check node health
- Check network latency to API server

### Problem: No Leader Elected

**Symptoms:**
```bash
kubectl logs -n kube-system -l app=kaws-operator | grep "leader"
# All pods: "attempting to acquire leader lease"
# None succeeding
```

**Causes:**
- Missing RBAC permissions
- Lease resource cannot be created
- API server connectivity issues

**Solutions:**
```bash
# Check RBAC
kubectl auth can-i create leases --as=system:serviceaccount:kube-system:kaws-operator -n kube-system

# Check if lease exists
kubectl get lease -n kube-system

# Manually delete lease if stuck
kubectl delete lease kaws-operator-lock -n kube-system
```

### Problem: Leader Not Reconciling

**Symptoms:**
- Leader elected successfully
- But no reconciliations happening

**Causes:**
- Informer cache not synced
- RBAC missing for watched resources
- EventRecycler CRD not created

**Solutions:**
```bash
# Check leader pod logs for cache sync
kubectl logs -n kube-system <leader-pod> | grep "cache"

# Check RBAC for events, pods, nodes
kubectl auth can-i list events --as=system:serviceaccount:kube-system:kaws-operator

# Ensure CRD is installed
kubectl get crd eventrecyclers.kaws.pischarti.dev
```

## References

- [controller-runtime Leader Election](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/leaderelection)
- [Kubernetes Lease Spec](https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/lease-v1/)
- [Leader Election Design](https://github.com/kubernetes/client-go/blob/master/tools/leaderelection/resourcelock/interface.go)

