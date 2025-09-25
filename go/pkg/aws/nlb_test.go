package aws

import (
	"testing"

	"github.com/pischarti/nix/go/pkg/vpc"
)

func TestParseNLBArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *vpc.NLBOptions
		wantErr  bool
	}{
		{
			name: "basic vpc only",
			args: []string{"nlb", "--vpc", "vpc-12345678"},
			expected: &vpc.NLBOptions{
				VPCID:  "vpc-12345678",
				Zone:   "",
				SortBy: "name",
			},
			wantErr: false,
		},
		{
			name: "vpc with zone",
			args: []string{"nlb", "--vpc", "vpc-12345678", "--zone", "us-east-1a"},
			expected: &vpc.NLBOptions{
				VPCID:  "vpc-12345678",
				Zone:   "us-east-1a",
				SortBy: "name",
			},
			wantErr: false,
		},
		{
			name: "vpc with zone and sort",
			args: []string{"nlb", "--vpc", "vpc-12345678", "--zone", "us-east-1a", "--sort", "state"},
			expected: &vpc.NLBOptions{
				VPCID:  "vpc-12345678",
				Zone:   "us-east-1a",
				SortBy: "state",
			},
			wantErr: false,
		},
		{
			name:     "invalid sort option",
			args:     []string{"nlb", "--vpc", "vpc-12345678", "--sort", "invalid"},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "no vpc provided",
			args: []string{"nlb", "--zone", "us-east-1a"},
			expected: &vpc.NLBOptions{
				VPCID:  "",
				Zone:   "us-east-1a",
				SortBy: "name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vpc.ParseNLBArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNLBArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != nil && tt.expected != nil {
				if result.VPCID != tt.expected.VPCID {
					t.Errorf("ParseNLBArgs() VPCID = %v, want %v", result.VPCID, tt.expected.VPCID)
				}
				if result.Zone != tt.expected.Zone {
					t.Errorf("ParseNLBArgs() Zone = %v, want %v", result.Zone, tt.expected.Zone)
				}
				if result.SortBy != tt.expected.SortBy {
					t.Errorf("ParseNLBArgs() SortBy = %v, want %v", result.SortBy, tt.expected.SortBy)
				}
			}
		})
	}
}

func TestSortNLBs(t *testing.T) {
	nlbs := []vpc.NLBInfo{
		{Name: "nlb-c", State: "active", Type: "network", Scheme: "internal"},
		{Name: "nlb-a", State: "active", Type: "network", Scheme: "external"},
		{Name: "nlb-b", State: "provisioning", Type: "network", Scheme: "internal"},
	}

	tests := []struct {
		name     string
		sortBy   string
		expected []string // expected order of names
	}{
		{
			name:     "sort by name",
			sortBy:   "name",
			expected: []string{"nlb-a", "nlb-b", "nlb-c"},
		},
		{
			name:     "sort by state",
			sortBy:   "state",
			expected: []string{"nlb-c", "nlb-a", "nlb-b"}, // active, active, provisioning
		},
		{
			name:     "sort by scheme",
			sortBy:   "scheme",
			expected: []string{"nlb-a", "nlb-c", "nlb-b"}, // external, internal, internal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the slice for each test
			testNLBs := make([]vpc.NLBInfo, len(nlbs))
			copy(testNLBs, nlbs)

			vpc.SortNLBs(testNLBs, tt.sortBy)

			// Check the order
			for i, expectedName := range tt.expected {
				if testNLBs[i].Name != expectedName {
					t.Errorf("SortNLBs() at index %d = %v, want %v", i, testNLBs[i].Name, expectedName)
				}
			}
		})
	}
}
