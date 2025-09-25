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
	"gopkg.in/yaml.v3"
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
			fmt.Println("Usage: aws ecr [--repository REPO_NAME] [--tag TAG] [--sort SORT_BY] [--all] [--older-than REFERENCE_TAG] [--output FORMAT]")
			fmt.Println("Options:")
			fmt.Println("  --repository REPO_NAME  ECR repository name (optional, use --all for all repos)")
			fmt.Println("  --tag TAG               Filter by image tag (optional)")
			fmt.Println("  --sort SORT_BY          Sort by: pushed (default), tag, size")
			fmt.Println("  --all                   List images from all repositories")
			fmt.Println("  --older-than REFERENCE_TAG  Show only images older than the reference tag")
			fmt.Println("  --output FORMAT         Output format: table (default), yaml")
			return nil, nil
		}
	}

	// Parse arguments
	opts, err := parseECRArgs(args)
	if err != nil {
		return nil, err
	}

	if opts.RepositoryName == "" && !opts.AllRepos {
		return nil, fmt.Errorf("repository parameter is required (use --repository REPO_NAME or --all for all repositories)")
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ECR client
	ecrClient := ecr.NewFromConfig(cfg)

	var images []ECRImageInfo

	if opts.AllRepos {
		// List all repositories first
		reposResult, err := ecrClient.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to describe repositories: %w", err)
		}

		// Get images from all repositories
		for _, repo := range reposResult.Repositories {
			input := &ecr.DescribeImagesInput{
				RepositoryName: repo.RepositoryName,
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
				// Log error but continue with other repositories
				fmt.Printf("Warning: failed to describe images in repository %s: %v\n", aws.ToString(repo.RepositoryName), err)
				continue
			}

			// Convert to ECRImageInfo structs and add to the list
			repoImages := convertECRImagesToImageInfo(result.ImageDetails)
			images = append(images, repoImages...)
		}
	} else {
		// Single repository
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
		images = convertECRImagesToImageInfo(result.ImageDetails)
	}

	// Filter images older than reference tag if specified
	var referenceDate *time.Time
	if opts.OlderThan != "" {
		images, referenceDate, err = filterImagesOlderThan(ecrClient, images, opts.OlderThan, opts.RepositoryName, opts.AllRepos)
		if err != nil {
			return nil, err
		}
	}

	// Sort images
	sortECRImages(images, opts.SortBy)

	// Print output in requested format
	switch opts.OutputFormat {
	case "yaml":
		printECRImagesYAML(images, opts, referenceDate)
	default:
		printECRImagesTable(images)
	}

	return nil, nil
}

// ECRArgs represents parsed ECR command arguments
type ECRArgs struct {
	RepositoryName string
	Tag            string
	SortBy         string
	AllRepos       bool
	OlderThan      string
	OutputFormat   string
}

// parseECRArgs parses command line arguments for ECR commands
func parseECRArgs(args []string) (*ECRArgs, error) {
	opts := &ECRArgs{
		SortBy:       "pushed", // default sort by push date (newest first)
		OutputFormat: "table",  // default output format
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
		case "--all":
			opts.AllRepos = true
		case "--older-than":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--older-than requires a value")
			}
			opts.OlderThan = args[i+1]
		case "--output":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			opts.OutputFormat = args[i+1]
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

// printECRImagesYAML prints ECR images in YAML format
func printECRImagesYAML(images []ECRImageInfo, opts *ECRArgs, referenceDate *time.Time) {
	// Convert to YAML-friendly structure
	yamlData := struct {
		Input struct {
			RepositoryName string     `yaml:"repository,omitempty"`
			Tag            string     `yaml:"tag,omitempty"`
			SortBy         string     `yaml:"sort_by,omitempty"`
			AllRepos       bool       `yaml:"all_repositories,omitempty"`
			OlderThan      string     `yaml:"older_than,omitempty"`
			OutputFormat   string     `yaml:"output_format,omitempty"`
			ReferenceDate  *time.Time `yaml:"reference_date,omitempty"`
		} `yaml:"input"`
		Images []ECRImageInfo `yaml:"images"`
		Count  int            `yaml:"count"`
	}{
		Input: struct {
			RepositoryName string     `yaml:"repository,omitempty"`
			Tag            string     `yaml:"tag,omitempty"`
			SortBy         string     `yaml:"sort_by,omitempty"`
			AllRepos       bool       `yaml:"all_repositories,omitempty"`
			OlderThan      string     `yaml:"older_than,omitempty"`
			OutputFormat   string     `yaml:"output_format,omitempty"`
			ReferenceDate  *time.Time `yaml:"reference_date,omitempty"`
		}{
			RepositoryName: opts.RepositoryName,
			Tag:            opts.Tag,
			SortBy:         opts.SortBy,
			AllRepos:       opts.AllRepos,
			OlderThan:      opts.OlderThan,
			OutputFormat:   opts.OutputFormat,
			ReferenceDate:  referenceDate,
		},
		Images: images,
		Count:  len(images),
	}

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		fmt.Printf("Error marshaling to YAML: %v\n", err)
		return
	}

	fmt.Print(string(yamlBytes))
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

// filterImagesOlderThan filters images to show only those older than the reference tag
func filterImagesOlderThan(ecrClient *ecr.Client, images []ECRImageInfo, referenceTag string, repositoryName string, allRepos bool) ([]ECRImageInfo, *time.Time, error) {
	var referenceTime *time.Time
	var err error

	if allRepos {
		// For all repositories, we need to find the reference tag across all repos
		referenceTime, err = findReferenceTagInAllRepos(ecrClient, referenceTag)
	} else {
		// For single repository, find the reference tag in that specific repo
		referenceTime, err = findReferenceTagInRepo(ecrClient, referenceTag, repositoryName)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to find reference tag '%s': %w", referenceTag, err)
	}

	if referenceTime == nil {
		fmt.Printf("Warning: Reference tag '%s' not found. Showing all images.\n", referenceTag)
		return images, nil, nil
	}

	// Filter images older than reference
	var filteredImages []ECRImageInfo
	for _, image := range images {
		if image.PushedAt.Before(*referenceTime) {
			filteredImages = append(filteredImages, image)
		}
	}

	return filteredImages, referenceTime, nil
}

// findReferenceTagInRepo finds the reference tag in a specific repository
func findReferenceTagInRepo(ecrClient *ecr.Client, referenceTag string, repositoryName string) (*time.Time, error) {
	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repositoryName),
		ImageIds: []types.ImageIdentifier{
			{
				ImageTag: aws.String(referenceTag),
			},
		},
	}

	result, err := ecrClient.DescribeImages(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	if len(result.ImageDetails) == 0 {
		return nil, nil // Tag not found
	}

	// Return the push time of the reference tag
	pushTime := aws.ToTime(result.ImageDetails[0].ImagePushedAt)
	return &pushTime, nil
}

// findReferenceTagInAllRepos finds the reference tag across all repositories
func findReferenceTagInAllRepos(ecrClient *ecr.Client, referenceTag string) (*time.Time, error) {
	// List all repositories
	reposResult, err := ecrClient.DescribeRepositories(context.TODO(), &ecr.DescribeRepositoriesInput{})
	if err != nil {
		return nil, err
	}

	// Search for the reference tag in each repository
	for _, repo := range reposResult.Repositories {
		input := &ecr.DescribeImagesInput{
			RepositoryName: repo.RepositoryName,
			ImageIds: []types.ImageIdentifier{
				{
					ImageTag: aws.String(referenceTag),
				},
			},
		}

		result, err := ecrClient.DescribeImages(context.TODO(), input)
		if err != nil {
			// Continue searching in other repositories
			continue
		}

		if len(result.ImageDetails) > 0 {
			// Found the reference tag, return its push time
			pushTime := aws.ToTime(result.ImageDetails[0].ImagePushedAt)
			return &pushTime, nil
		}
	}

	return nil, nil // Tag not found in any repository
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
