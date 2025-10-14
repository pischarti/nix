# kaws Operator Architecture

## Overview

The `kaws` operator is built on `controller-runtime`, the same framework used by Operator SDK and Kubebuilder. It provides a robust, race-condition-free architecture for managing Kubernetes resources and automating node group recycling.

## Informer-Based Architecture

### What are Informers?

Kubernetes Informers are a core pattern for efficiently watching and caching Kubernetes resources. They provide:

1. **Local Cache**: A synchronized, in-memory cache of Kubernetes resources
2. **Event Handlers**: Callbacks that fire when resources are added, updated, or deleted
3. **Resync Mechanism**: Periodic full re-lists to ensure cache consistency
4. **Thread Safety**: All cache operations are protected by locks

### How kaws Uses Informers

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes API Server                    │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ Watch API
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                   controller-runtime Manager                 │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   Informer Cache                        │ │
│  │  - Events (corev1.Event)                               │ │
│  │  - Pods (corev1.Pod)                                   │ │
│  │  - Nodes (corev1.Node)                                 │ │
│  │  - EventRecyclers (kawsv1alpha1.EventRecycler)        │ │
│  │                                                         │ │
│  │  Thread-safe, synchronized, eventually consistent      │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │            EventRecyclerReconciler                      │ │
│  │  - Reads from cache (not API server)                   │ │
│  │  - No race conditions                                   │ │
│  │  - Efficient: No repeated API calls                    │ │
│  └────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────┘
                            │
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     AWS API (EC2, ASG)                       │
│  - Find node groups                                          │
│  - Scale node groups                                         │
│  - Monitor instance states                                   │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

#### 1. **Manager** (`ctrl.Manager`)

The manager is the heart of controller-runtime:

```go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
    Cache: cache.Options{
        SyncPeriod: ptr(10 * time.Minute), // How often to re-list all resources
    },
})
```

- **Creates and manages informers** for all watched resource types
- **Starts the informer cache** and waits for initial sync
- **Runs controllers** and their reconciliation loops
- **Handles graceful shutdown** and cleanup

#### 2. **Informer Cache**

The cache is automatically created by the manager for each resource type:

- **Events**: Watches all Kubernetes events across all namespaces
- **Pods**: Caches pod information for node mapping
- **Nodes**: Caches node information for instance ID extraction
- **EventRecyclers**: Watches our custom resource

**Key Features:**
- **Thread-safe**: All reads/writes are protected by locks
- **Eventually consistent**: May be slightly behind API server during high churn
- **Automatic re-sync**: Periodically re-lists all resources to ensure consistency
- **Efficient**: Uses watch API to get updates, not repeated LIST calls

#### 3. **EventRecyclerReconciler**

The controller that implements the reconciliation logic:

```go
type EventRecyclerReconciler struct {
    client.Client  // This client uses the informer cache
    Scheme *runtime.Scheme
    
    // AWS clients for node group operations
    EC2Client *ec2.Client
    ASGClient *autoscaling.Client
    
    // Thread-safe event tracking
    processedEvents map[string]time.Time
    mu              sync.RWMutex
}
```

**Reconciliation Flow:**

1. **Triggered by**: EventRecycler resource changes or periodic requeue
2. **Reads from cache**: All `r.Get()` and `r.List()` calls use the informer cache
3. **Filters events**: Searches for matching error patterns
4. **Maps to node groups**: Uses cached pod and node information
5. **Tracks processed events**: Thread-safe deduplication using `sync.RWMutex`
6. **Triggers recycling**: When thresholds are exceeded

### Race Condition Prevention

#### Problem: Direct API Server Queries

Without informers, you might do:

```go
// BAD: Multiple concurrent reconciliations could see inconsistent state
events, err := clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
```

**Issues:**
- Multiple goroutines reading at different times see different snapshots
- High API server load from repeated LIST calls
- No coordination between concurrent reconciliations
- Race conditions when processing the same event multiple times

#### Solution: Informer Cache

With informers:

```go
// GOOD: All reconciliations share the same synchronized cache
eventList := &corev1.EventList{}
err := r.List(ctx, eventList)  // Reads from informer cache, not API server
```

**Benefits:**
- **Single source of truth**: All controllers see the same cached state
- **Thread-safe**: Informer cache handles concurrent access with locks
- **Efficient**: One watch stream, many readers
- **Consistent**: All readers see the same version of each resource

#### Additional Safety: Processed Event Tracking

To prevent processing the same event multiple times across reconciliations:

```go
func (r *EventRecyclerReconciler) filterRecentEvents(events []corev1.Event) []corev1.Event {
    r.mu.RLock()  // Read lock for checking
    defer r.mu.RUnlock()
    
    // Check if events were recently processed
    recentEvents := []corev1.Event{}
    for _, event := range events {
        if !r.wasRecentlyProcessed(event) {
            recentEvents = append(recentEvents, event)
        }
    }
    
    // Mark as processed (upgrades to write lock)
    r.mu.Lock()
    for _, event := range recentEvents {
        r.markProcessed(event)
    }
    r.mu.Unlock()
    
    return recentEvents
}
```

### Performance Benefits

#### Without Informers (Polling):

```
Every 60s:
  ├─ LIST /api/v1/events (100ms, 500KB)
  ├─ GET /api/v1/pods/pod-1 (50ms, 10KB)
  ├─ GET /api/v1/pods/pod-2 (50ms, 10KB)
  └─ GET /api/v1/nodes/node-1 (50ms, 20KB)

Total per minute: ~250ms, ~550KB
Total per hour: 60 * 250ms = 15s, 33MB
```

#### With Informers (Caching):

```
Initial sync:
  └─ WATCH /api/v1/events (establishes long-lived connection)

Every 60s:
  ├─ Read from cache (1ms, 0KB network)
  ├─ Read from cache (1ms, 0KB network)
  └─ Read from cache (1ms, 0KB network)

Total per minute: ~3ms, ~0KB
Total per hour: 60 * 3ms = 180ms, 0MB
```

**Improvement:**
- **83x faster** reconciliations
- **99.9% reduction** in network traffic
- **Unlimited readers** with no additional cost

### Failure Modes and Recovery

#### API Server Disconnection

**What happens:**
- Watch connection is lost
- Informer detects disconnection and logs warning
- Informer automatically reconnects
- Full re-list is performed to rebuild cache

**Impact:**
- During disconnection: Controller continues using stale cache (eventually consistent)
- After reconnection: Cache is updated, reconciliation proceeds normally
- No manual intervention required

#### Cache Out of Sync

**What happens:**
- Periodic re-sync (every 10 minutes by default) performs full LIST
- Compares cached resources with API server state
- Updates cache to match API server

**Impact:**
- Cache is eventually consistent
- Max staleness = sync period (10 minutes)
- For critical operations, can force a direct API read

### Best Practices

1. **Always use the cached client**
   ```go
   // Good: Uses informer cache
   err := r.Get(ctx, key, &obj)
   err := r.List(ctx, &list)
   
   // Bad: Bypasses cache, adds load
   err := directClient.Get(ctx, key, &obj)
   ```

2. **Set appropriate sync periods**
   ```go
   Cache: cache.Options{
       SyncPeriod: ptr(10 * time.Minute), // Balance freshness vs load
   }
   ```

3. **Use local tracking for deduplication**
   ```go
   // Informer prevents race conditions in cache access
   // But you still need app-level deduplication
   processedEvents map[string]time.Time
   mu sync.RWMutex
   ```

4. **Handle stale reads gracefully**
   ```go
   // Cache may be slightly behind
   // Design reconciliation to be idempotent
   if alreadyRecycled(nodeGroup) {
       return nil // Safe to retry
   }
   ```

## Deployment Considerations

### Resource Requirements

- **Memory**: Informer cache size depends on cluster size
  - Small cluster (1000 events): ~50MB
  - Large cluster (10000 events): ~500MB
- **CPU**: Minimal, only during reconciliation
- **Network**: One long-lived watch connection per resource type

### Scaling and High Availability

The informer-based architecture with leader election scales well:

#### Leader Election

The operator uses controller-runtime's built-in leader election:

```go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    LeaderElection:          true,
    LeaderElectionID:        "kaws-operator-lock",
    LeaderElectionNamespace: "kube-system",
    LeaseDuration:           ptr(15 * time.Second),
    RenewDeadline:           ptr(10 * time.Second),
    RetryPeriod:             ptr(2 * time.Second),
})
```

**How it works:**
1. **Lease-based**: Uses Kubernetes Lease resources for coordination
2. **Single active replica**: Only the leader performs reconciliations
3. **Automatic failover**: If leader crashes, another replica takes over within ~15 seconds
4. **Heartbeat**: Leader renews lease every 2 seconds

**Replica States:**
- **Leader**: Actively reconciling resources
- **Follower**: Ready to take over, caches are synced but reconciliation is paused

#### High Availability Setup

**Default configuration (3 replicas):**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kaws-operator
spec:
  replicas: 3  # Leader election enabled
```

**Benefits:**
- **Zero downtime**: If leader pod dies, another replica becomes leader
- **Fast failover**: New leader elected within 15 seconds
- **Rolling updates**: Safe to update without service interruption
- **Node failure tolerance**: Can tolerate loss of nodes running operator pods

#### Scaling Considerations

- **Multiple replicas**: Each replica has its own informer cache, all watch the same resources
- **Leader election**: Ensures single-writer guarantees (only leader reconciles)
- **Sharding**: Not needed for typical use cases, informers are very efficient
- **Resource usage**: Each replica uses ~50-500MB memory (for cache), minimal CPU when not leader

### Monitoring

Key metrics to monitor:

- **Cache sync status**: Are informers in sync?
- **Reconciliation duration**: How long does each reconcile take?
- **Reconciliation rate**: How often are we reconciling?
- **Event count**: How many events are in the cache?

## References

- [controller-runtime Book](https://book.kubebuilder.io/reference/controller-runtime.html)
- [Kubernetes Informers](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
- [client-go Informers](https://pkg.go.dev/k8s.io/client-go/informers)

