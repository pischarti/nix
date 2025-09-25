package vpc

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// ParseSubnetsArgs parses command line arguments for the subnets command
func ParseSubnetsArgs(args []string) (*SubnetsOptions, error) {
	opts := &SubnetsOptions{
		SortBy: "cidr", // Default sort by CIDR
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--vpc":
			if i+1 < len(args) {
				i++
				opts.VPCID = args[i]
			}
		case "--zone":
			if i+1 < len(args) {
				i++
				opts.Zone = args[i]
			}
		case "--sort":
			if i+1 < len(args) {
				i++
				opts.SortBy = args[i]
			}
		}
	}

	// Validate sort option
	validSorts := map[string]bool{"cidr": true, "az": true, "name": true, "type": true}
	if !validSorts[opts.SortBy] {
		return nil, fmt.Errorf("invalid sort option '%s'. Valid options: cidr, az, name, type", opts.SortBy)
	}

	return opts, nil
}

// ParseNLBArgs parses command line arguments for the nlb command
func ParseNLBArgs(args []string) (*NLBOptions, error) {
	opts := &NLBOptions{
		SortBy: "name", // Default sort by name
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--vpc":
			if i+1 < len(args) {
				i++
				opts.VPCID = args[i]
			}
		case "--zone":
			if i+1 < len(args) {
				i++
				opts.Zone = args[i]
			}
		case "--sort":
			if i+1 < len(args) {
				i++
				opts.SortBy = args[i]
			}
		}
	}

	// Validate sort option
	validSorts := map[string]bool{"name": true, "state": true, "type": true, "scheme": true, "created": true}
	if !validSorts[opts.SortBy] {
		return nil, fmt.Errorf("invalid sort option '%s'. Valid options: name, state, type, scheme, created", opts.SortBy)
	}

	return opts, nil
}

// SortSubnets sorts a slice of SubnetInfo based on the specified sort criteria
func SortSubnets(subnets []SubnetInfo, sortBy string) {
	switch sortBy {
	case "cidr":
		sort.Slice(subnets, func(i, j int) bool {
			return CompareCIDRBlocks(subnets[i].CIDRBlock, subnets[j].CIDRBlock) < 0
		})
	case "az":
		sort.Slice(subnets, func(i, j int) bool {
			return subnets[i].AZ < subnets[j].AZ
		})
	case "name":
		sort.Slice(subnets, func(i, j int) bool {
			return subnets[i].Name < subnets[j].Name
		})
	case "type":
		sort.Slice(subnets, func(i, j int) bool {
			return subnets[i].Type < subnets[j].Type
		})
	}
}

// CompareCIDRBlocks compares two CIDR blocks for sorting
func CompareCIDRBlocks(cidr1, cidr2 string) int {
	_, ipNet1, err1 := net.ParseCIDR(cidr1)
	_, ipNet2, err2 := net.ParseCIDR(cidr2)

	if err1 != nil || err2 != nil {
		// If parsing fails, fall back to string comparison
		return strings.Compare(cidr1, cidr2)
	}

	// Compare network addresses
	network1 := ipNet1.IP
	network2 := ipNet2.IP

	// Convert to bytes for comparison
	bytes1 := network1.To4()
	bytes2 := network2.To4()

	if bytes1 == nil || bytes2 == nil {
		// IPv6 or invalid addresses, fall back to string comparison
		return strings.Compare(cidr1, cidr2)
	}

	// Compare byte by byte
	for i := 0; i < 4; i++ {
		if bytes1[i] < bytes2[i] {
			return -1
		} else if bytes1[i] > bytes2[i] {
			return 1
		}
	}

	// If network addresses are the same, compare prefix lengths
	prefix1, _ := ipNet1.Mask.Size()
	prefix2, _ := ipNet2.Mask.Size()

	if prefix1 < prefix2 {
		return -1
	} else if prefix1 > prefix2 {
		return 1
	}

	return 0
}

// ConvertEC2SubnetsToSubnetInfo converts AWS EC2 subnet types to SubnetInfo structs
func ConvertEC2SubnetsToSubnetInfo(ec2Subnets []types.Subnet) []SubnetInfo {
	var subnets []SubnetInfo

	for _, subnet := range ec2Subnets {
		name := ""
		subnetType := "subnet"
		var relevantTags []string

		// Extract name and type from tags, and collect other relevant tags
		for _, tag := range subnet.Tags {
			key := aws.ToString(tag.Key)
			value := aws.ToString(tag.Value)

			switch key {
			case "Name":
				name = value
			case "Type":
				subnetType = value
			default:
				// Include tags that are commonly used for networking/infrastructure
				if strings.HasPrefix(key, "kubernetes.io/") ||
					strings.HasPrefix(key, "aws:") ||
					strings.HasPrefix(key, "Name") ||
					key == "Environment" ||
					key == "Project" ||
					key == "Tier" {
					relevantTags = append(relevantTags, key)
				}
			}
		}

		// Format tags with each tag on a separate line
		tagsStr := strings.Join(relevantTags, "\n")

		subnetInfo := SubnetInfo{
			SubnetID:  aws.ToString(subnet.SubnetId),
			CIDRBlock: aws.ToString(subnet.CidrBlock),
			AZ:        aws.ToString(subnet.AvailabilityZone),
			Name:      name,
			State:     string(subnet.State),
			Type:      subnetType,
			Tags:      tagsStr,
		}
		subnets = append(subnets, subnetInfo)
	}

	return subnets
}

// SortNLBs sorts a slice of NLBInfo based on the specified sort criteria
func SortNLBs(nlbs []NLBInfo, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(nlbs, func(i, j int) bool {
			return nlbs[i].Name < nlbs[j].Name
		})
	case "state":
		sort.Slice(nlbs, func(i, j int) bool {
			return nlbs[i].State < nlbs[j].State
		})
	case "type":
		sort.Slice(nlbs, func(i, j int) bool {
			return nlbs[i].Type < nlbs[j].Type
		})
	case "scheme":
		sort.Slice(nlbs, func(i, j int) bool {
			return nlbs[i].Scheme < nlbs[j].Scheme
		})
	case "created":
		sort.Slice(nlbs, func(i, j int) bool {
			return nlbs[i].CreatedTime < nlbs[j].CreatedTime
		})
	}
}
