package aws

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/jedib0t/go-pretty/v6/table"
	"gofr.dev/pkg/gofr"
)

// ECRImageInfo represents ECR image information
type ECRImageInfo struct {
	RepositoryName string
	ImageTag       string
	ImageDigest    string
	PushedAt       time.Time
	ImageSize      int64
	ImageManifest  string
}

// ListECRImages handles the ecr command for listing AWS ECR images
func ListECRImages(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check for help flag first
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: aws ecr --repository REPO_NAME [--tag TAG] [--sort SORT_BY]")
			fmt.Println("Options:")
			fmt.Println("  --repository REPO_NAME  ECR repository name (required)")
			fmt.Println("  --tag TAG               Filter by image tag (optional)")
			fmt.Println("  --sort SORT_BY          Sort by: pushed (default), tag, size")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := parseECRArgs(args)
	if err != nil {
		return nil, err
	}

	if opts.RepositoryName == "" {
		return nil, fmt.Errorf("repository parameter is required")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)

	// Describe images
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(opts.RepositoryName),
	}

	if opts.Tag != "" {
		input.ImageIds = []types.ImageIdentifier{
			{
				ImageTag: aws.String(opts.Tag),
			},
		}
	}

	result, err := ecrClient.DescribeImages(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	// Convert to ECRImageInfo structs
	images := convertECRImagesToImageInfo(result.ImageDetails)

	// Sort images
	sortECRImages(images, opts.SortBy)

	// Print table output
	printECRImagesTable(images)

	return nil, nil
}

// ECRArgs represents parsed ECR command arguments
type ECRArgs struct {
	RepositoryName string
	Tag            string
	SortBy         string
}

// parseECRArgs parses command line arguments for ECR commands
func parseECRArgs(args []string) (*ECRArgs, error) {
	opts := &ECRArgs{
		SortBy: "pushed", // default sort by push date (newest first)
	}

	for i, arg := range args {
		switch arg {
		case "--repository":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--repository requires a value")
			}
			opts.RepositoryName = args[i+1]
		case "--tag":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--tag requires a value")
			}
			opts.Tag = args[i+1]
		case "--sort":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--sort requires a value")
			}
			opts.SortBy = args[i+1]
		}
	}

	return opts, nil
}

// convertECRImagesToImageInfo converts ECR image details to ECRImageInfo structs
func convertECRImagesToImageInfo(imageDetails []types.ImageDetail) []ECRImageInfo {
	var images []ECRImageInfo

	for _, detail := range imageDetails {
		// Handle multiple tags per image
		if len(detail.ImageTags) == 0 {
			// Image without tags
			images = append(images, ECRImageInfo{
				RepositoryName: aws.ToString(detail.RepositoryName),
				ImageTag:       "<untagged>",
				ImageDigest:    aws.ToString(detail.ImageDigest),
				PushedAt:       aws.ToTime(detail.ImagePushedAt),
				ImageSize:      aws.ToInt64(detail.ImageSizeInBytes),
				ImageManifest:  aws.ToString(detail.ImageManifestMediaType),
			})
		} else {
			// Create an entry for each tag
			for _, tag := range detail.ImageTags {
				images = append(images, ECRImageInfo{
					RepositoryName: aws.ToString(detail.RepositoryName),
					ImageTag:       tag,
					ImageDigest:    aws.ToString(detail.ImageDigest),
					PushedAt:       aws.ToTime(detail.ImagePushedAt),
					ImageSize:      aws.ToInt64(detail.ImageSizeInBytes),
					ImageManifest:  aws.ToString(detail.ImageManifestMediaType),
				})
			}
		}
	}

	return images
}

// sortECRImages sorts ECR images based on the specified criteria
func sortECRImages(images []ECRImageInfo, sortBy string) {
	switch sortBy {
	case "pushed":
		sort.Slice(images, func(i, j int) bool {
			return images[i].PushedAt.After(images[j].PushedAt)
		})
	case "size":
		sort.Slice(images, func(i, j int) bool {
			return images[i].ImageSize > images[j].ImageSize
		})
	case "tag":
		fallthrough
	default:
		sort.Slice(images, func(i, j int) bool {
			return images[i].ImageTag < images[j].ImageTag
		})
	}
}

// printECRImagesTable prints ECR images in a formatted table
func printECRImagesTable(images []ECRImageInfo) {
	if len(images) == 0 {
		fmt.Println("No images found in the repository.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)
	t.AppendHeader(table.Row{"Repository", "Tag", "Digest", "Pushed At", "Size", "Manifest"})

	// Set column widths to keep tag column narrow
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Tag", WidthMax: 20}, // Limit tag column to 20 characters
	})

	for _, image := range images {
		// Format size in human-readable format
		sizeStr := formatBytes(image.ImageSize)

		// Truncate digest for display
		digest := image.ImageDigest
		if len(digest) > 12 {
			digest = digest[:12] + "..."
		}

		// Format pushed time
		pushedStr := image.PushedAt.Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			image.RepositoryName,
			image.ImageTag,
			digest,
			pushedStr,
			sizeStr,
			image.ImageManifest,
		})
	}

	t.Render()
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ECRRouter handles ECR command routing
func ECRRouter(ctx *gofr.Context) (any, error) {
	args := os.Args[1:] // Get command line args for parsing flags

	// Check if we have a subcommand (not a flag)
	if len(args) >= 2 {
		subcommand := args[1]
		// Only treat as subcommand if it's not a flag (doesn't start with --)
		if !strings.HasPrefix(subcommand, "--") {
			switch subcommand {
			case "list":
				return ListECRImages(ctx)
			default:
				return nil, fmt.Errorf("unknown ECR subcommand: %s. Use 'aws ecr --help' for usage information", subcommand)
			}
		}
	}

	// Default to list command (handles both no args and flags)
	return ListECRImages(ctx)
}
