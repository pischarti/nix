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
