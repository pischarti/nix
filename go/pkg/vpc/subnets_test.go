package vpc

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestParseSubnetsArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expected    *SubnetsOptions
		expectError bool
	}{
		{
			name: "valid args with vpc only",
			args: []string{"--vpc", "vpc-12345678"},
			expected: &SubnetsOptions{
				VPCID:  "vpc-12345678",
				Zone:   "",
				SortBy: "cidr",
			},
			expectError: false,
		},
		{
			name: "valid args with vpc and zone",
			args: []string{"--vpc", "vpc-12345678", "--zone", "us-east-1a"},
			expected: &SubnetsOptions{
				VPCID:  "vpc-12345678",
				Zone:   "us-east-1a",
				SortBy: "cidr",
			},
			expectError: false,
		},
		{
			name: "valid args with all options",
			args: []string{"--vpc", "vpc-12345678", "--zone", "us-east-1a", "--sort", "az"},
			expected: &SubnetsOptions{
				VPCID:  "vpc-12345678",
				Zone:   "us-east-1a",
				SortBy: "az",
			},
			expectError: false,
		},
		{
			name: "valid sort options",
			args: []string{"--vpc", "vpc-12345678", "--sort", "name"},
			expected: &SubnetsOptions{
				VPCID:  "vpc-12345678",
				Zone:   "",
				SortBy: "name",
			},
			expectError: false,
		},
		{
			name: "valid sort options - type",
			args: []string{"--vpc", "vpc-12345678", "--sort", "type"},
			expected: &SubnetsOptions{
				VPCID:  "vpc-12345678",
				Zone:   "",
				SortBy: "type",
			},
			expectError: false,
		},
		{
			name:        "invalid sort option",
			args:        []string{"--vpc", "vpc-12345678", "--sort", "invalid"},
			expected:    nil,
			expectError: true,
		},
		{
			name: "empty args",
			args: []string{},
			expected: &SubnetsOptions{
				VPCID:  "",
				Zone:   "",
				SortBy: "cidr",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSubnetsArgs(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.VPCID != tt.expected.VPCID {
				t.Errorf("VPCID = %v, want %v", result.VPCID, tt.expected.VPCID)
			}
			if result.Zone != tt.expected.Zone {
				t.Errorf("Zone = %v, want %v", result.Zone, tt.expected.Zone)
			}
			if result.SortBy != tt.expected.SortBy {
				t.Errorf("SortBy = %v, want %v", result.SortBy, tt.expected.SortBy)
			}
		})
	}
}

func TestSortSubnets(t *testing.T) {
	tests := []struct {
		name     string
		subnets  []SubnetInfo
		sortBy   string
		expected []SubnetInfo
	}{
		{
			name: "sort by cidr",
			subnets: []SubnetInfo{
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "subnet2"},
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "subnet1"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "subnet3"},
			},
			sortBy: "cidr",
			expected: []SubnetInfo{
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "subnet1"},
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "subnet2"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "subnet3"},
			},
		},
		{
			name: "sort by az",
			subnets: []SubnetInfo{
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1c", Name: "subnet1"},
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1a", Name: "subnet2"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1b", Name: "subnet3"},
			},
			sortBy: "az",
			expected: []SubnetInfo{
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1a", Name: "subnet2"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1b", Name: "subnet3"},
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1c", Name: "subnet1"},
			},
		},
		{
			name: "sort by name",
			subnets: []SubnetInfo{
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "zebra"},
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "alpha"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "beta"},
			},
			sortBy: "name",
			expected: []SubnetInfo{
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "alpha"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "beta"},
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "zebra"},
			},
		},
		{
			name: "sort by type",
			subnets: []SubnetInfo{
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "subnet1", Type: "private"},
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "subnet2", Type: "public"},
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "subnet3", Type: "database"},
			},
			sortBy: "type",
			expected: []SubnetInfo{
				{CIDRBlock: "10.0.3.0/24", AZ: "us-east-1c", Name: "subnet3", Type: "database"},
				{CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", Name: "subnet1", Type: "private"},
				{CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", Name: "subnet2", Type: "public"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of subnets to avoid modifying the original
			subnets := make([]SubnetInfo, len(tt.subnets))
			copy(subnets, tt.subnets)

			SortSubnets(subnets, tt.sortBy)

			if len(subnets) != len(tt.expected) {
				t.Errorf("Length mismatch: got %d, want %d", len(subnets), len(tt.expected))
				return
			}

			for i, subnet := range subnets {
				if subnet.CIDRBlock != tt.expected[i].CIDRBlock {
					t.Errorf("CIDRBlock[%d] = %v, want %v", i, subnet.CIDRBlock, tt.expected[i].CIDRBlock)
				}
				if subnet.AZ != tt.expected[i].AZ {
					t.Errorf("AZ[%d] = %v, want %v", i, subnet.AZ, tt.expected[i].AZ)
				}
				if subnet.Name != tt.expected[i].Name {
					t.Errorf("Name[%d] = %v, want %v", i, subnet.Name, tt.expected[i].Name)
				}
				if subnet.Type != tt.expected[i].Type {
					t.Errorf("Type[%d] = %v, want %v", i, subnet.Type, tt.expected[i].Type)
				}
			}
		})
	}
}

func TestCompareCIDRBlocks(t *testing.T) {
	tests := []struct {
		name     string
		cidr1    string
		cidr2    string
		expected int
	}{
		{
			name:     "same CIDR blocks",
			cidr1:    "10.0.1.0/24",
			cidr2:    "10.0.1.0/24",
			expected: 0,
		},
		{
			name:     "different networks, cidr1 smaller",
			cidr1:    "10.0.1.0/24",
			cidr2:    "10.0.2.0/24",
			expected: -1,
		},
		{
			name:     "different networks, cidr1 larger",
			cidr1:    "10.0.2.0/24",
			cidr2:    "10.0.1.0/24",
			expected: 1,
		},
		{
			name:     "same network, different prefix lengths",
			cidr1:    "10.0.1.0/24",
			cidr2:    "10.0.1.0/16",
			expected: 1, // /24 is larger than /16
		},
		{
			name:     "same network, different prefix lengths reversed",
			cidr1:    "10.0.1.0/16",
			cidr2:    "10.0.1.0/24",
			expected: -1, // /16 is smaller than /24
		},
		{
			name:     "invalid CIDR blocks",
			cidr1:    "invalid",
			cidr2:    "10.0.1.0/24",
			expected: 1, // string comparison: "invalid" > "10.0.1.0/24"
		},
		{
			name:     "both invalid CIDR blocks",
			cidr1:    "invalid1",
			cidr2:    "invalid2",
			expected: -1, // string comparison: "invalid1" < "invalid2"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareCIDRBlocks(tt.cidr1, tt.cidr2)
			if result != tt.expected {
				t.Errorf("CompareCIDRBlocks(%v, %v) = %v, want %v", tt.cidr1, tt.cidr2, result, tt.expected)
			}
		})
	}
}

func TestConvertEC2SubnetsToSubnetInfo(t *testing.T) {
	tests := []struct {
		name       string
		ec2Subnets []types.Subnet
		expected   []SubnetInfo
	}{
		{
			name: "basic conversion",
			ec2Subnets: []types.Subnet{
				{
					SubnetId:         aws.String("subnet-12345678"),
					VpcId:            aws.String("vpc-12345678"),
					CidrBlock:        aws.String("10.0.1.0/24"),
					AvailabilityZone: aws.String("us-east-1a"),
					State:            types.SubnetStateAvailable,
					Tags: []types.Tag{
						{Key: aws.String("Name"), Value: aws.String("test-subnet")},
						{Key: aws.String("Type"), Value: aws.String("private")},
					},
				},
			},
			expected: []SubnetInfo{
				{
					SubnetID:  "subnet-12345678",
					VPCID:     "vpc-12345678",
					CIDRBlock: "10.0.1.0/24",
					AZ:        "us-east-1a",
					Name:      "test-subnet",
					State:     "available",
					Type:      "private",
				},
			},
		},
		{
			name: "conversion without tags",
			ec2Subnets: []types.Subnet{
				{
					SubnetId:         aws.String("subnet-87654321"),
					VpcId:            aws.String("vpc-87654321"),
					CidrBlock:        aws.String("10.0.2.0/24"),
					AvailabilityZone: aws.String("us-east-1b"),
					State:            types.SubnetStatePending,
					Tags:             []types.Tag{},
				},
			},
			expected: []SubnetInfo{
				{
					SubnetID:  "subnet-87654321",
					VPCID:     "vpc-87654321",
					CIDRBlock: "10.0.2.0/24",
					AZ:        "us-east-1b",
					Name:      "",
					State:     "pending",
					Type:      "subnet",
				},
			},
		},
		{
			name: "conversion with partial tags",
			ec2Subnets: []types.Subnet{
				{
					SubnetId:         aws.String("subnet-11111111"),
					VpcId:            aws.String("vpc-11111111"),
					CidrBlock:        aws.String("10.0.3.0/24"),
					AvailabilityZone: aws.String("us-east-1c"),
					State:            types.SubnetStateAvailable,
					Tags: []types.Tag{
						{Key: aws.String("Name"), Value: aws.String("named-subnet")},
						{Key: aws.String("Environment"), Value: aws.String("prod")},
					},
				},
			},
			expected: []SubnetInfo{
				{
					SubnetID:  "subnet-11111111",
					VPCID:     "vpc-11111111",
					CIDRBlock: "10.0.3.0/24",
					AZ:        "us-east-1c",
					Name:      "named-subnet",
					State:     "available",
					Type:      "subnet",
				},
			},
		},
		{
			name:       "empty subnets list",
			ec2Subnets: []types.Subnet{},
			expected:   []SubnetInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertEC2SubnetsToSubnetInfo(tt.ec2Subnets)

			if len(result) != len(tt.expected) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, subnet := range result {
				expected := tt.expected[i]
				if subnet.SubnetID != expected.SubnetID {
					t.Errorf("SubnetID[%d] = %v, want %v", i, subnet.SubnetID, expected.SubnetID)
				}
				if subnet.VPCID != expected.VPCID {
					t.Errorf("VPCID[%d] = %v, want %v", i, subnet.VPCID, expected.VPCID)
				}
				if subnet.CIDRBlock != expected.CIDRBlock {
					t.Errorf("CIDRBlock[%d] = %v, want %v", i, subnet.CIDRBlock, expected.CIDRBlock)
				}
				if subnet.AZ != expected.AZ {
					t.Errorf("AZ[%d] = %v, want %v", i, subnet.AZ, expected.AZ)
				}
				if subnet.Name != expected.Name {
					t.Errorf("Name[%d] = %v, want %v", i, subnet.Name, expected.Name)
				}
				if subnet.State != expected.State {
					t.Errorf("State[%d] = %v, want %v", i, subnet.State, expected.State)
				}
				if subnet.Type != expected.Type {
					t.Errorf("Type[%d] = %v, want %v", i, subnet.Type, expected.Type)
				}
			}
		})
	}
}
