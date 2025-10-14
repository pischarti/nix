package vpc

// SubnetInfo represents information about an AWS subnet
type SubnetInfo struct {
	SubnetID  string
	CIDRBlock string
	AZ        string
	Name      string
	State     string
	Type      string
	Tags      string
}

// SubnetsOptions represents the parsed command line options for the subnets command
type SubnetsOptions struct {
	VPCID  string
	Zone   string
	SortBy string
}

// NLBInfo represents information about an AWS Network Load Balancer
type NLBInfo struct {
	LoadBalancerArn   string
	Name              string
	DNSName           string
	State             string
	Type              string
	Scheme            string
	VPCID             string
	AvailabilityZones string
	Subnets           string
	CreatedTime       string
	Tags              string
}

// NLBOptions represents the parsed command line options for the nlb command
type NLBOptions struct {
	VPCID  string
	Zone   string
	SortBy string
}
