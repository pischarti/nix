//go:build integration
// +build integration

package operator

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MockEC2Client is a mock implementation of EC2 client for testing
type MockEC2Client struct {
	instances map[string]*ec2types.Instance
}

func NewMockEC2Client() *MockEC2Client {
	return &MockEC2Client{
		instances: make(map[string]*ec2types.Instance),
	}
}

// AddInstance adds a mock instance
func (m *MockEC2Client) AddInstance(instanceID, nodeGroup, cluster string) {
	m.instances[instanceID] = &ec2types.Instance{
		InstanceId:   aws.String(instanceID),
		InstanceType: ec2types.InstanceTypeT3Medium,
		State: &ec2types.InstanceState{
			Name: ec2types.InstanceStateNameRunning,
		},
		Tags: []ec2types.Tag{
			{Key: aws.String("eks:nodegroup-name"), Value: aws.String(nodeGroup)},
			{Key: aws.String("eks:cluster-name"), Value: aws.String(cluster)},
		},
	}
}

// DescribeInstances mocks the EC2 DescribeInstances API
func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	var instances []ec2types.Instance

	for _, instanceID := range params.InstanceIds {
		if instance, found := m.instances[instanceID]; found {
			instances = append(instances, *instance)
		}
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{
			{
				Instances: instances,
			},
		},
	}, nil
}

// MockASGClient is a mock implementation of Auto Scaling client
type MockASGClient struct {
	asgs map[string]*asgtypes.AutoScalingGroup
}

func NewMockASGClient() *MockASGClient {
	return &MockASGClient{
		asgs: make(map[string]*asgtypes.AutoScalingGroup),
	}
}

// AddAutoScalingGroup adds a mock ASG
func (m *MockASGClient) AddAutoScalingGroup(name string, minSize, maxSize, desiredCapacity int32) {
	m.asgs[name] = &asgtypes.AutoScalingGroup{
		AutoScalingGroupName: aws.String(name),
		MinSize:              aws.Int32(minSize),
		MaxSize:              aws.Int32(maxSize),
		DesiredCapacity:      aws.Int32(desiredCapacity),
		Tags: []asgtypes.TagDescription{
			{Key: aws.String("eks:nodegroup-name"), Value: aws.String(name)},
		},
	}
}

// DescribeAutoScalingGroups mocks the ASG DescribeAutoScalingGroups API
func (m *MockASGClient) DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	var asgs []asgtypes.AutoScalingGroup

	if len(params.AutoScalingGroupNames) == 0 {
		// Return all ASGs
		for _, asg := range m.asgs {
			asgs = append(asgs, *asg)
		}
	} else {
		// Return specific ASGs
		for _, name := range params.AutoScalingGroupNames {
			if asg, found := m.asgs[name]; found {
				asgs = append(asgs, *asg)
			}
		}
	}

	return &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: asgs,
	}, nil
}

// UpdateAutoScalingGroup mocks the ASG UpdateAutoScalingGroup API
func (m *MockASGClient) UpdateAutoScalingGroup(ctx context.Context, params *autoscaling.UpdateAutoScalingGroupInput, optFns ...func(*autoscaling.Options)) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	name := *params.AutoScalingGroupName
	if asg, found := m.asgs[name]; found {
		if params.MinSize != nil {
			asg.MinSize = params.MinSize
		}
		if params.MaxSize != nil {
			asg.MaxSize = params.MaxSize
		}
		if params.DesiredCapacity != nil {
			asg.DesiredCapacity = params.DesiredCapacity
		}
	}

	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

// TestMockNodeGroupRecycling tests recycling with pure mocks (no localstack needed)
func TestMockNodeGroupRecycling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test doesn't require localstack - uses pure mocks
	t.Log("Testing node group recycling with mock AWS clients...")

	// Setup mock AWS clients
	mockEC2 := NewMockEC2Client()
	mockASG := NewMockASGClient()

	// Add mock instance matching our test node
	mockEC2.AddInstance("i-1234567890abcdef0", "test-nodegroup-1", "test-cluster")

	// Add mock ASG
	mockASG.AddAutoScalingGroup("test-nodegroup-1", 1, 3, 2)

	ctx := context.Background()

	// Test: Describe instances
	describeResult, err := mockEC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{"i-1234567890abcdef0"},
	})
	if err != nil {
		t.Fatalf("Mock DescribeInstances failed: %v", err)
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		t.Fatal("Expected to find mock instance")
	}

	instance := describeResult.Reservations[0].Instances[0]
	t.Logf("Found instance: %s", *instance.InstanceId)

	// Verify tags
	var nodeGroupName string
	for _, tag := range instance.Tags {
		if *tag.Key == "eks:nodegroup-name" {
			nodeGroupName = *tag.Value
			break
		}
	}

	if nodeGroupName != "test-nodegroup-1" {
		t.Errorf("Expected node group test-nodegroup-1, got %s", nodeGroupName)
	}

	// Test: Describe ASG
	asgResult, err := mockASG.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{"test-nodegroup-1"},
	})
	if err != nil {
		t.Fatalf("Mock DescribeAutoScalingGroups failed: %v", err)
	}

	if len(asgResult.AutoScalingGroups) == 0 {
		t.Fatal("Expected to find mock ASG")
	}

	asg := asgResult.AutoScalingGroups[0]
	t.Logf("Found ASG: %s (Min: %d, Max: %d, Desired: %d)",
		*asg.AutoScalingGroupName, *asg.MinSize, *asg.MaxSize, *asg.DesiredCapacity)

	// Verify ASG configuration
	if *asg.MinSize != 1 {
		t.Errorf("Expected MinSize=1, got %d", *asg.MinSize)
	}
	if *asg.MaxSize != 3 {
		t.Errorf("Expected MaxSize=3, got %d", *asg.MaxSize)
	}
	if *asg.DesiredCapacity != 2 {
		t.Errorf("Expected DesiredCapacity=2, got %d", *asg.DesiredCapacity)
	}

	// Test: Update ASG (simulate scaling down)
	_, err = mockASG.UpdateAutoScalingGroup(ctx, &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("test-nodegroup-1"),
		MinSize:              aws.Int32(0),
		MaxSize:              aws.Int32(0),
		DesiredCapacity:      aws.Int32(0),
	})
	if err != nil {
		t.Fatalf("Mock UpdateAutoScalingGroup failed: %v", err)
	}

	// Verify ASG was updated
	asgResult2, _ := mockASG.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{"test-nodegroup-1"},
	})

	asg2 := asgResult2.AutoScalingGroups[0]
	if *asg2.MinSize != 0 || *asg2.MaxSize != 0 || *asg2.DesiredCapacity != 0 {
		t.Errorf("Expected ASG to be scaled to 0, got Min=%d, Max=%d, Desired=%d",
			*asg2.MinSize, *asg2.MaxSize, *asg2.DesiredCapacity)
	}

	t.Log("✅ Mock ASG scaling verified - recycling logic can be tested")

	// Test: Restore original values (simulate scaling up)
	_, err = mockASG.UpdateAutoScalingGroup(ctx, &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String("test-nodegroup-1"),
		MinSize:              aws.Int32(1),
		MaxSize:              aws.Int32(3),
		DesiredCapacity:      aws.Int32(2),
	})
	if err != nil {
		t.Fatalf("Mock UpdateAutoScalingGroup (restore) failed: %v", err)
	}

	// Verify restored
	asgResult3, _ := mockASG.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{"test-nodegroup-1"},
	})

	asg3 := asgResult3.AutoScalingGroups[0]
	if *asg3.DesiredCapacity != 2 {
		t.Errorf("Expected restored DesiredCapacity=2, got %d", *asg3.DesiredCapacity)
	}

	t.Log("✅ Mock ASG restore verified - complete recycle simulation successful")
}
