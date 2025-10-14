package print

import (
	"testing"

	"github.com/pischarti/nix/go/pkg/vpc"
)

func TestPrintNLBTable(t *testing.T) {
	// Test with empty slice
	nlbs := []vpc.NLBInfo{}
	PrintNLBTable(nlbs) // Should not panic and should print "No Network Load Balancers found."

	// Test with sample data
	nlbs = []vpc.NLBInfo{
		{
			LoadBalancerArn:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/test-nlb/1234567890123456",
			Name:              "test-nlb",
			DNSName:           "test-nlb-1234567890.us-east-1.elb.amazonaws.com",
			State:             "active",
			Type:              "network",
			Scheme:            "internal",
			VPCID:             "vpc-12345678",
			AvailabilityZones: "us-east-1a, us-east-1b",
			Subnets:           "subnet-12345678, subnet-87654321",
			CreatedTime:       "2024-01-01T12:00:00Z",
			Tags:              "Environment=prod\nProject=test",
		},
		{
			LoadBalancerArn:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/test-nlb-2/1234567890123457",
			Name:              "test-nlb-2",
			DNSName:           "test-nlb-2-1234567890.us-east-1.elb.amazonaws.com",
			State:             "provisioning",
			Type:              "network",
			Scheme:            "external",
			VPCID:             "vpc-12345678",
			AvailabilityZones: "us-east-1a",
			Subnets:           "subnet-12345678",
			CreatedTime:       "2024-01-02T12:00:00Z",
			Tags:              "Environment=dev",
		},
		{
			LoadBalancerArn:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/test-nlb-3/1234567890123458",
			Name:              "", // Empty name to test fallback
			DNSName:           "test-nlb-3-1234567890.us-east-1.elb.amazonaws.com",
			State:             "active",
			Type:              "network",
			Scheme:            "internal",
			VPCID:             "vpc-12345678",
			AvailabilityZones: "us-east-1b",
			Subnets:           "subnet-87654321",
			CreatedTime:       "2024-01-03T12:00:00Z",
			Tags:              "Environment=prod",
		},
	}

	// This test mainly ensures the function doesn't panic
	// In a real test environment, you might want to capture stdout
	PrintNLBTable(nlbs)
}

func TestNLBInfoFields(t *testing.T) {
	nlb := vpc.NLBInfo{
		LoadBalancerArn:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/test-nlb/1234567890123456",
		Name:              "test-nlb",
		DNSName:           "test-nlb-1234567890.us-east-1.elb.amazonaws.com",
		State:             "active",
		Type:              "network",
		Scheme:            "internal",
		VPCID:             "vpc-12345678",
		AvailabilityZones: "us-east-1a, us-east-1b",
		Subnets:           "subnet-12345678, subnet-87654321",
		CreatedTime:       "2024-01-01T12:00:00Z",
		Tags:              "Environment=prod\nProject=test",
	}

	// Test that all fields are populated
	if nlb.LoadBalancerArn == "" {
		t.Error("LoadBalancerArn should not be empty")
	}
	if nlb.Name == "" {
		t.Error("Name should not be empty")
	}
	if nlb.DNSName == "" {
		t.Error("DNSName should not be empty")
	}
	if nlb.State == "" {
		t.Error("State should not be empty")
	}
	if nlb.Type == "" {
		t.Error("Type should not be empty")
	}
	if nlb.Scheme == "" {
		t.Error("Scheme should not be empty")
	}
	if nlb.VPCID == "" {
		t.Error("VPCID should not be empty")
	}
	if nlb.AvailabilityZones == "" {
		t.Error("AvailabilityZones should not be empty")
	}
	if nlb.Subnets == "" {
		t.Error("Subnets should not be empty")
	}
	if nlb.CreatedTime == "" {
		t.Error("CreatedTime should not be empty")
	}
}
