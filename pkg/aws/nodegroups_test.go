package aws

import (
	"testing"
)

func TestNodeGroupInfo_Struct(t *testing.T) {
	// Test that NodeGroupInfo can be created with all fields
	info := NodeGroupInfo{
		InstanceID:    "i-1234567890abcdef0",
		NodeGroupName: "ng-workers-1",
		ClusterName:   "my-cluster",
		Status:        "active",
		InstanceType:  "t3.large",
	}

	if info.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %q, want %q", info.InstanceID, "i-1234567890abcdef0")
	}
	if info.NodeGroupName != "ng-workers-1" {
		t.Errorf("NodeGroupName = %q, want %q", info.NodeGroupName, "ng-workers-1")
	}
	if info.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %q, want %q", info.ClusterName, "my-cluster")
	}
	if info.Status != "active" {
		t.Errorf("Status = %q, want %q", info.Status, "active")
	}
	if info.InstanceType != "t3.large" {
		t.Errorf("InstanceType = %q, want %q", info.InstanceType, "t3.large")
	}
}

// Note: Testing FindNodeGroups would require either:
// 1. Mock AWS clients (using interfaces)
// 2. AWS SDK fake/stub clients
// 3. Integration tests with real AWS
// These are typically done in integration tests rather than unit tests
