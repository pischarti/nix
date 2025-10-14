package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// NodeGroupInfo contains information about a node group and its instances
type NodeGroupInfo struct {
	InstanceID    string
	NodeGroupName string
	ClusterName   string
	Status        string
	InstanceType  string
}

// FindNodeGroups queries AWS to find node groups for the given instance IDs
func FindNodeGroups(ctx context.Context, ec2Client *ec2.Client, instanceIDs []string, clusterName string) ([]NodeGroupInfo, error) {
	results := []NodeGroupInfo{}

	// Get instance details from EC2
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}

	describeOutput, err := ec2Client.DescribeInstances(ctx, describeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	// Build a map of instance ID to instance details
	instanceMap := make(map[string]struct {
		InstanceType string
		Tags         map[string]string
	})

	for _, reservation := range describeOutput.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId == nil {
				continue
			}

			tags := make(map[string]string)
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}

			instanceType := ""
			if instance.InstanceType != "" {
				instanceType = string(instance.InstanceType)
			}

			instanceMap[*instance.InstanceId] = struct {
				InstanceType string
				Tags         map[string]string
			}{
				InstanceType: instanceType,
				Tags:         tags,
			}
		}
	}

	// Extract node group information from tags
	for instanceID, details := range instanceMap {
		info := NodeGroupInfo{
			InstanceID:   instanceID,
			InstanceType: details.InstanceType,
			Status:       "N/A",
		}

		// Try to get cluster name from tags
		if cluster, ok := details.Tags["eks:cluster-name"]; ok {
			info.ClusterName = cluster
		} else if cluster, ok := details.Tags["kubernetes.io/cluster/"+clusterName]; ok {
			info.ClusterName = strings.TrimPrefix(cluster, "kubernetes.io/cluster/")
		}

		// Try to get node group name from tags
		if nodeGroup, ok := details.Tags["eks:nodegroup-name"]; ok {
			info.NodeGroupName = nodeGroup
		} else if nodeGroup, ok := details.Tags["alpha.eksctl.io/nodegroup-name"]; ok {
			info.NodeGroupName = nodeGroup
		}

		// If we still don't have node group info, mark as unknown
		if info.NodeGroupName == "" {
			info.NodeGroupName = "Unknown"
		}
		if info.ClusterName == "" {
			info.ClusterName = "Unknown"
		}

		results = append(results, info)
	}

	return results, nil
}
