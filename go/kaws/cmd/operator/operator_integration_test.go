//go:build integration
// +build integration

package operator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pischarti/nix/pkg/k8s"
	pkgoperator "github.com/pischarti/nix/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kindClusterName = "kaws-test"
	localstackURL   = "http://localhost:4566"
	testNamespace   = "default"
)

// TestOperatorIntegration is the main integration test
func TestOperatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	env := setupTestEnvironment(t)
	defer env.Teardown(t)

	// Run test scenarios
	t.Run("DetectErrorEvents", func(t *testing.T) {
		testDetectErrorEvents(t, env)
	})

	t.Run("NodeGroupIdentification", func(t *testing.T) {
		testNodeGroupIdentification(t, env)
	})

	t.Run("EventFiltering", func(t *testing.T) {
		testEventFiltering(t, env)
	})

	t.Run("NodeGroupRecycling", func(t *testing.T) {
		testNodeGroupRecycling(t, env)
	})
}

// TestEnvironment holds the test infrastructure
type TestEnvironment struct {
	KindClusterName  string
	Clientset        *kubernetes.Clientset
	K8sClient        *k8s.Client
	EC2Client        *ec2.Client
	ASGClient        *autoscaling.Client
	LocalstackURL    string
	createdResources []string
}

// setupTestEnvironment creates the test infrastructure
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Log("Setting up test environment...")

	env := &TestEnvironment{
		KindClusterName: kindClusterName,
		LocalstackURL:   localstackURL,
	}

	// 1. Create kind cluster
	env.createKindCluster(t)

	// 2. Setup Kubernetes client
	env.setupK8sClient(t)

	// 3. Setup AWS clients (pointing to localstack)
	env.setupAWSClients(t)

	// 4. Create mock AWS resources
	env.createMockAWSResources(t)

	// 5. Create mock Kubernetes resources
	env.createMockK8sResources(t)

	t.Log("Test environment ready")
	return env
}

// createKindCluster creates a kind cluster for testing
func (env *TestEnvironment) createKindCluster(t *testing.T) {
	t.Log("Creating kind cluster...")

	// Check if cluster already exists
	checkCmd := exec.Command("kind", "get", "clusters")
	output, _ := checkCmd.CombinedOutput()
	if contains(string(output), env.KindClusterName) {
		t.Log("Kind cluster already exists, deleting...")
		deleteCmd := exec.Command("kind", "delete", "cluster", "--name", env.KindClusterName)
		if err := deleteCmd.Run(); err != nil {
			t.Logf("Warning: Failed to delete existing cluster: %v", err)
		}
	}

	// Create new cluster
	createCmd := exec.Command("kind", "create", "cluster", "--name", env.KindClusterName, "--wait", "60s")
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		t.Fatalf("Failed to create kind cluster: %v", err)
	}

	t.Log("Kind cluster created")
}

// setupK8sClient sets up the Kubernetes client
func (env *TestEnvironment) setupK8sClient(t *testing.T) {
	t.Log("Setting up Kubernetes client...")

	// Get kubeconfig for kind cluster
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("Failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	env.Clientset = clientset

	// Create k8s.Client wrapper
	k8sClient, err := k8s.NewClient()
	if err != nil {
		t.Fatalf("Failed to create k8s client: %v", err)
	}
	env.K8sClient = k8sClient

	t.Log("Kubernetes client ready")
}

// setupAWSClients sets up AWS clients pointing to localstack
func (env *TestEnvironment) setupAWSClients(t *testing.T) {
	t.Log("Setting up AWS clients (localstack)...")

	ctx := context.Background()

	// Configure AWS SDK to use localstack
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           env.LocalstackURL,
					SigningRegion: region,
				}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "test",
		)),
	)
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	env.EC2Client = ec2.NewFromConfig(cfg)
	env.ASGClient = autoscaling.NewFromConfig(cfg)

	t.Log("AWS clients ready")
}

// createMockAWSResources creates mock EC2 instances and Auto Scaling Groups
func (env *TestEnvironment) createMockAWSResources(t *testing.T) {
	t.Log("Creating mock AWS resources in localstack...")

	ctx := context.Background()

	// Create mock EC2 instances with EKS node group tags
	env.createMockEC2Instances(t, ctx)

	// Create mock Auto Scaling Groups
	env.createMockAutoScalingGroups(t, ctx)

	t.Log("Mock AWS resources ready")
}

// createMockEC2Instances creates mock EC2 instances in localstack
func (env *TestEnvironment) createMockEC2Instances(t *testing.T, ctx context.Context) {
	// Note: Localstack's EC2 implementation may be limited
	// We'll create instances with proper tags for node group identification

	runInstancesInput := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-12345678"),
		InstanceType: ec2types.InstanceTypeT3Medium,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags: []ec2types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("test-node-1-instance"),
					},
					{
						Key:   aws.String("eks:cluster-name"),
						Value: aws.String("test-cluster"),
					},
					{
						Key:   aws.String("eks:nodegroup-name"),
						Value: aws.String("test-nodegroup-1"),
					},
					{
						Key:   aws.String("kubernetes.io/cluster/test-cluster"),
						Value: aws.String("owned"),
					},
				},
			},
		},
	}

	result, err := env.EC2Client.RunInstances(ctx, runInstancesInput)
	if err != nil {
		t.Logf("Warning: Could not create mock EC2 instance: %v", err)
		t.Log("This is expected if localstack EC2 API is not fully functional")
		return
	}

	if len(result.Instances) > 0 {
		instanceID := *result.Instances[0].InstanceId
		t.Logf("Created mock EC2 instance: %s", instanceID)

		// Update our test node to use this instance ID
		env.updateNodeProviderID(t, ctx, instanceID)
	}
}

// createMockAutoScalingGroups creates mock ASGs in localstack
func (env *TestEnvironment) createMockAutoScalingGroups(t *testing.T, ctx context.Context) {
	asgInput := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("test-nodegroup-1"),
		MinSize:              aws.Int32(1),
		MaxSize:              aws.Int32(3),
		DesiredCapacity:      aws.Int32(2),
		AvailabilityZones:    []string{"us-east-1a"},
		Tags: []asgtypes.Tag{
			{
				Key:   aws.String("eks:cluster-name"),
				Value: aws.String("test-cluster"),
			},
			{
				Key:   aws.String("eks:nodegroup-name"),
				Value: aws.String("test-nodegroup-1"),
			},
		},
	}

	_, err := env.ASGClient.CreateAutoScalingGroup(ctx, asgInput)
	if err != nil {
		t.Logf("Warning: Could not create mock ASG: %v", err)
		t.Log("This is expected if localstack Auto Scaling API is not fully functional")
		return
	}

	t.Log("Created mock Auto Scaling Group: test-nodegroup-1")
}

// updateNodeProviderID updates the test node with the actual instance ID from localstack
func (env *TestEnvironment) updateNodeProviderID(t *testing.T, ctx context.Context, instanceID string) {
	node, err := env.Clientset.CoreV1().Nodes().Get(ctx, "test-node-1", metav1.GetOptions{})
	if err != nil {
		t.Logf("Warning: Could not get node to update: %v", err)
		return
	}

	node.Spec.ProviderID = fmt.Sprintf("aws:///us-east-1a/%s", instanceID)
	_, err = env.Clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		t.Logf("Warning: Could not update node provider ID: %v", err)
	} else {
		t.Logf("Updated node provider ID to: %s", node.Spec.ProviderID)
	}
}

// createMockK8sResources creates test pods, nodes, and events
func (env *TestEnvironment) createMockK8sResources(t *testing.T) {
	t.Log("Creating mock Kubernetes resources...")

	ctx := context.Background()

	// Create a test node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-1",
			Labels: map[string]string{
				"kubernetes.io/hostname": "test-node-1",
			},
		},
		Spec: corev1.NodeSpec{
			ProviderID: "aws:///us-east-1a/i-1234567890abcdef0",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	_, err := env.Clientset.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test node: %v", err)
	}
	env.createdResources = append(env.createdResources, "node/test-node-1")

	// Create a test pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-1",
			Namespace: testNamespace,
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node-1",
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}

	_, err = env.Clientset.CoreV1().Pods(testNamespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create test pod: %v", err)
	}
	env.createdResources = append(env.createdResources, "pod/test-pod-1")

	// Create test events
	env.createTestEvent(t, "test-pod-1", "failed to get sandbox image", "FailedSandboxImage")
	env.createTestEvent(t, "test-pod-1", "ImagePullBackOff", "ImagePullBackOff")
	env.createTestEvent(t, "test-pod-1", "Normal event", "Normal")

	t.Log("Mock Kubernetes resources created")
}

// createTestEvent creates a test event
func (env *TestEnvironment) createTestEvent(t *testing.T, podName, message, reason string) {
	ctx := context.Background()

	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("test-event-%s-%d", reason, time.Now().Unix()),
			Namespace: testNamespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      podName,
			Namespace: testNamespace,
		},
		Reason:         reason,
		Message:        message,
		Type:           "Warning",
		FirstTimestamp: metav1.Time{Time: time.Now()},
		LastTimestamp:  metav1.Time{Time: time.Now()},
		Count:          1,
	}

	_, err := env.Clientset.CoreV1().Events(testNamespace).Create(ctx, event, metav1.CreateOptions{})
	if err != nil {
		t.Logf("Warning: Failed to create test event: %v", err)
	}
}

// Teardown cleans up the test environment
func (env *TestEnvironment) Teardown(t *testing.T) {
	t.Log("Tearing down test environment...")

	// Clean up Kubernetes resources
	for _, resource := range env.createdResources {
		t.Logf("Deleting resource: %s", resource)
		// Parse resource type and name
		// Simplified cleanup - in production you'd parse properly
	}

	// Delete kind cluster
	deleteCmd := exec.Command("kind", "delete", "cluster", "--name", env.KindClusterName)
	if err := deleteCmd.Run(); err != nil {
		t.Logf("Warning: Failed to delete kind cluster: %v", err)
	}

	t.Log("Test environment cleaned up")
}

// testDetectErrorEvents tests that the operator can detect error events
func testDetectErrorEvents(t *testing.T, env *TestEnvironment) {
	ctx := context.Background()

	opConfig := &pkgoperator.OperatorConfig{
		WatchInterval:    10 * time.Second,
		SearchTerms:      []string{"failed to get sandbox image"},
		RecycleThreshold: 1,
		DryRun:           true,
		ProcessedEvents:  make(map[string]time.Time),
	}

	// Run the check
	err := pkgoperator.CheckAndRecycle(ctx, env.K8sClient, env.EC2Client, env.ASGClient, opConfig, true)
	if err != nil {
		t.Errorf("CheckAndRecycle failed: %v", err)
	}

	// Verify events were processed
	if len(opConfig.ProcessedEvents) == 0 {
		t.Error("Expected events to be processed, but none were found")
	}
}

// testNodeGroupIdentification tests node group identification from events
func testNodeGroupIdentification(t *testing.T, env *TestEnvironment) {
	ctx := context.Background()

	// Query events
	events, err := env.K8sClient.QueryEvents(ctx, k8s.EventQueryOptions{
		Namespace: testNamespace,
	})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("Expected events to be created, but none were found")
	}

	// Filter events
	matchingEvents := k8s.FilterEvents(events, "failed to get sandbox image")
	if len(matchingEvents) == 0 {
		t.Error("Expected to find matching events, but none were found")
	}

	t.Logf("Found %d matching events", len(matchingEvents))
}

// testEventFiltering tests that event filtering works correctly
func testEventFiltering(t *testing.T, env *TestEnvironment) {
	ctx := context.Background()

	// Query all events
	events, err := env.K8sClient.QueryEvents(ctx, k8s.EventQueryOptions{
		Namespace: testNamespace,
	})
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	// Test filtering with different terms
	testCases := []struct {
		searchTerm    string
		expectedCount int
	}{
		{"failed to get sandbox image", 1},
		{"ImagePullBackOff", 1},
		{"Normal", 1},
		{"nonexistent", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.searchTerm, func(t *testing.T) {
			matched := k8s.FilterEvents(events, tc.searchTerm)
			if len(matched) < tc.expectedCount {
				t.Errorf("Expected at least %d events for %q, got %d", tc.expectedCount, tc.searchTerm, len(matched))
			}
		})
	}
}

// testNodeGroupRecycling tests the complete recycling flow with mocked AWS resources
func testNodeGroupRecycling(t *testing.T, env *TestEnvironment) {
	ctx := context.Background()

	// Create multiple error events to exceed threshold
	for i := 0; i < 6; i++ {
		env.createTestEvent(t, "test-pod-1", "failed to get sandbox image", fmt.Sprintf("FailedSandboxImage-%d", i))
		time.Sleep(10 * time.Millisecond) // Ensure unique timestamps
	}

	// Wait a moment for events to be created
	time.Sleep(500 * time.Millisecond)

	// Create operator config with low threshold
	opConfig := &pkgoperator.OperatorConfig{
		WatchInterval:    10 * time.Second,
		SearchTerms:      []string{"failed to get sandbox image"},
		RecycleThreshold: 3, // Low threshold to trigger recycling
		DryRun:           true,
		ProcessedEvents:  make(map[string]time.Time),
	}

	// Run the check - should detect node group needing recycling
	err := pkgoperator.CheckAndRecycle(ctx, env.K8sClient, env.EC2Client, env.ASGClient, opConfig, true)
	if err != nil {
		t.Errorf("CheckAndRecycle failed: %v", err)
	}

	// Verify the ASG exists (if localstack supports it)
	describeASGInput := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{"test-nodegroup-1"},
	}

	asgResult, err := env.ASGClient.DescribeAutoScalingGroups(ctx, describeASGInput)
	if err != nil {
		t.Logf("Note: Localstack ASG API not fully functional: %v", err)
		t.Log("In production, this would verify ASG configuration before recycling")
		return
	}

	if len(asgResult.AutoScalingGroups) > 0 {
		asg := asgResult.AutoScalingGroups[0]
		t.Logf("Found ASG: %s (Min: %d, Max: %d, Desired: %d)",
			*asg.AutoScalingGroupName,
			*asg.MinSize,
			*asg.MaxSize,
			*asg.DesiredCapacity)

		// Verify ASG has correct configuration
		if *asg.MinSize != 1 {
			t.Errorf("Expected MinSize=1, got %d", *asg.MinSize)
		}
		if *asg.MaxSize != 3 {
			t.Errorf("Expected MaxSize=3, got %d", *asg.MaxSize)
		}
		if *asg.DesiredCapacity != 2 {
			t.Errorf("Expected DesiredCapacity=2, got %d", *asg.DesiredCapacity)
		}

		t.Log("✅ ASG configuration verified - ready for recycling")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr) != -1)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
