package print

import (
	"strings"
	"testing"

	"github.com/pischarti/nix/pkg/vpc"
)

func TestPrintSubnetsTableString(t *testing.T) {
	tests := []struct {
		name     string
		subnets  []vpc.SubnetInfo
		expected []string // Expected strings to be present in output
	}{
		{
			name: "single subnet with tags",
			subnets: []vpc.SubnetInfo{
				{
					SubnetID:  "subnet-12345678",
					CIDRBlock: "10.0.1.0/24",
					AZ:        "us-east-1a",
					Name:      "test-subnet",
					State:     "available",
					Type:      "private",
					Tags:      "kubernetes.io/role/elb\nEnvironment",
				},
			},
			expected: []string{
				"SUBNET ID",
				"CIDR BLOCK",
				"AZ",
				"NAME",
				"STATE",
				"TYPE",
				"TAGS",
				"subnet-12345678",
				"10.0.1.0/24",
				"us-east-1a",
				"test-subnet",
				"available",
				"private",
				"kubernetes.io/role/elb",
				"Environment",
			},
		},
		{
			name: "multiple subnets with different tags",
			subnets: []vpc.SubnetInfo{
				{
					SubnetID:  "subnet-11111111",
					CIDRBlock: "10.0.1.0/24",
					AZ:        "us-east-1a",
					Name:      "subnet-1",
					State:     "available",
					Type:      "private",
					Tags:      "kubernetes.io/role/elb",
				},
				{
					SubnetID:  "subnet-22222222",
					CIDRBlock: "10.0.2.0/24",
					AZ:        "us-east-1b",
					Name:      "subnet-2",
					State:     "available",
					Type:      "public",
					Tags:      "kubernetes.io/cluster/mycluster\nEnvironment",
				},
			},
			expected: []string{
				"SUBNET ID",
				"CIDR BLOCK",
				"AZ",
				"NAME",
				"STATE",
				"TYPE",
				"TAGS",
				"subnet-11111111",
				"subnet-22222222",
				"10.0.1.0/24",
				"10.0.2.0/24",
				"us-east-1a",
				"us-east-1b",
				"subnet-1",
				"subnet-2",
				"available",
				"private",
				"public",
				"kubernetes.io/role/elb",
				"kubernetes.io/cluster/mycluster",
				"Environment",
			},
		},
		{
			name:    "empty subnets list",
			subnets: []vpc.SubnetInfo{},
			expected: []string{
				"SUBNET ID",
				"CIDR BLOCK",
				"AZ",
				"NAME",
				"STATE",
				"TYPE",
				"TAGS",
			},
		},
		{
			name: "subnet with empty name and no tags",
			subnets: []vpc.SubnetInfo{
				{
					SubnetID:  "subnet-33333333",
					CIDRBlock: "10.0.3.0/24",
					AZ:        "us-east-1c",
					Name:      "",
					State:     "pending",
					Type:      "subnet",
					Tags:      "",
				},
			},
			expected: []string{
				"SUBNET ID",
				"CIDR BLOCK",
				"AZ",
				"NAME",
				"STATE",
				"TYPE",
				"TAGS",
				"subnet-33333333",
				"10.0.3.0/24",
				"us-east-1c",
				"pending",
				"subnet",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrintSubnetsTableString(tt.subnets)

			// Check that result is not empty
			if result == "" {
				t.Error("Expected non-empty table output")
				return
			}

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected table output to contain %q, but it didn't. Output: %s", expected, result)
				}
			}

			// Check that the output contains table structure (spaces for alignment)
			hasTableStructure := strings.Contains(result, "  ") // Multiple spaces indicate table formatting
			if !hasTableStructure {
				t.Errorf("Expected table output to contain table formatting, but it didn't. Output: %s", result)
			}
		})
	}
}

func TestPrintSubnetsTableStringOutput(t *testing.T) {
	// Test that the function returns a string and doesn't panic
	subnets := []vpc.SubnetInfo{
		{
			SubnetID:  "subnet-test",
			CIDRBlock: "10.0.0.0/24",
			AZ:        "us-east-1a",
			Name:      "test",
			State:     "available",
			Type:      "private",
			Tags:      "kubernetes.io/role/elb",
		},
	}

	result := PrintSubnetsTableString(subnets)

	// Should not be empty
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Should contain headers
	if !strings.Contains(result, "SUBNET ID") {
		t.Error("Expected result to contain 'SUBNET ID' header")
	}

	// Should contain TAGS header
	if !strings.Contains(result, "TAGS") {
		t.Error("Expected result to contain 'TAGS' header")
	}

	// Should contain data
	if !strings.Contains(result, "subnet-test") {
		t.Error("Expected result to contain subnet data")
	}

	// Should contain tags
	if !strings.Contains(result, "kubernetes.io/role/elb") {
		t.Error("Expected result to contain tag data")
	}
}

func TestPrintSubnetsTableStringWithSpecialCharacters(t *testing.T) {
	// Test with special characters in names, types, and tags
	subnets := []vpc.SubnetInfo{
		{
			SubnetID:  "subnet-12345678",
			CIDRBlock: "10.0.1.0/24",
			AZ:        "us-east-1a",
			Name:      "subnet-with-special-chars-!@#$%",
			State:     "available",
			Type:      "private-subnet",
			Tags:      "kubernetes.io/role/elb, aws:autoscaling:groupName",
		},
	}

	result := PrintSubnetsTableString(subnets)

	// Should not be empty and should handle special characters
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Should contain the special characters
	if !strings.Contains(result, "subnet-with-special-chars-!@#$%") {
		t.Error("Expected result to contain special characters in name")
	}

	if !strings.Contains(result, "private-subnet") {
		t.Error("Expected result to contain hyphen in type")
	}

	// Should contain AWS tags
	if !strings.Contains(result, "aws:autoscaling:groupName") {
		t.Error("Expected result to contain AWS tag")
	}
}
