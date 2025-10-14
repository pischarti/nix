package print

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pischarti/nix/pkg/vpc"
)

// PrintNLBTable prints NLBs in a table format
func PrintNLBTable(nlbs []vpc.NLBInfo) {
	if len(nlbs) == 0 {
		fmt.Println("No Network Load Balancers found.")
		return
	}

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)

	// Set table headers
	t.AppendHeader(table.Row{
		"Name",
		"State",
		"Scheme",
		"AZ / Subnet",
		"Created Time",
		"Tags",
	})

	// Add rows
	for _, nlb := range nlbs {
		// Truncate long values for better display
		// Use name from tag, fallback to Load Balancer ARN if name is empty
		name := nlb.Name
		if name == "" {
			name = nlb.LoadBalancerArn
			if len(name) > 20 {
				name = name[:17] + "..."
			}
		}

		// Format availability zones and subnets to show matching pairs on separate lines
		azs := formatAZSubnetPairs(nlb.AvailabilityZones, nlb.Subnets)

		createdTime := nlb.CreatedTime
		if len(createdTime) > 19 {
			// Format to show just date and time without timezone
			parts := strings.Split(createdTime, "T")
			if len(parts) >= 2 {
				timePart := strings.Split(parts[1], ".")[0]
				createdTime = parts[0] + " " + timePart
			}
		}

		// Format tags - show first few tags, truncate if too many
		tags := nlb.Tags
		if tags != "" {
			tagLines := strings.Split(tags, "\n")
			if len(tagLines) > 3 {
				tags = strings.Join(tagLines[:3], "\n") + "\n..."
			}
		}

		t.AppendRow(table.Row{
			name,
			nlb.State,
			nlb.Scheme,
			azs,
			createdTime,
			tags,
		})
	}

	// Configure table options
	t.SetAutoIndex(false)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 20}, // Name
		{Number: 2, WidthMax: 10}, // State
		{Number: 3, WidthMax: 10}, // Scheme
		{Number: 4, WidthMax: 50}, // AZ / Subnet
		{Number: 5, WidthMax: 19}, // Created Time
		{Number: 6, WidthMax: 30}, // Tags
	})

	// Render table
	t.Render()

	// Print summary
	fmt.Printf("\nFound %d Network Load Balancer(s)\n", len(nlbs))
}

// formatAZSubnetPairs formats availability zones and subnets to show matching pairs on separate lines
func formatAZSubnetPairs(azs, subnets string) string {
	if azs == "" || subnets == "" {
		return ""
	}

	// Split the comma-separated values
	azList := strings.Split(azs, ", ")
	subnetList := strings.Split(subnets, ", ")

	// Trim whitespace from each item
	for i, az := range azList {
		azList[i] = strings.TrimSpace(az)
	}
	for i, subnet := range subnetList {
		subnetList[i] = strings.TrimSpace(subnet)
	}

	// Create pairs, handling cases where counts might not match
	var pairs []string
	maxLen := len(azList)
	if len(subnetList) > maxLen {
		maxLen = len(subnetList)
	}

	for i := 0; i < maxLen; i++ {
		var az, subnet string
		if i < len(azList) {
			az = azList[i]
		}
		if i < len(subnetList) {
			subnet = subnetList[i]
		}

		if az != "" && subnet != "" {
			pairs = append(pairs, az+" / "+subnet)
		} else if az != "" {
			pairs = append(pairs, az)
		} else if subnet != "" {
			pairs = append(pairs, subnet)
		}
	}

	return strings.Join(pairs, "\n")
}
