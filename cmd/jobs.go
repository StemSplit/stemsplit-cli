package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/StemSplit/stemsplit-cli/internal/client"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List recent stem separation jobs",
	Long:  `List recent stem separation jobs with their status, file, and timing.`,
	RunE:  runJobs,
}

var jobsLimit int

func init() {
	jobsCmd.Flags().IntVarP(&jobsLimit, "limit", "n", 10, "number of jobs to show (max 100)")
}

func runJobs(cmd *cobra.Command, args []string) error {
	apiKey, err := resolveAPIKey()
	if err != nil {
		return err
	}

	c := client.New(apiKey)
	resp, err := c.ListJobs(jobsLimit, 0)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(resp.Jobs) == 0 {
		fmt.Println("No jobs found. Run: stemsplit separate <file>")
		return nil
	}

	// Header
	fmt.Printf("\n%s%-26s  %-12s  %-10s  %-20s  %s%s\n",
		colorBold,
		"JOB ID", "STATUS", "TYPE", "CREATED", "FILE",
		colorReset,
	)
	fmt.Println(strings.Repeat("─", 90))

	for _, job := range resp.Jobs {
		created := parseTime(job.CreatedAt)
		filename := job.Input.FileName
		if len(filename) > 30 {
			filename = filename[:27] + "..."
		}
		outputType := strings.ToLower(strings.ReplaceAll(job.Options.OutputType, "_", "-"))

		fmt.Printf("%-26s  %-22s  %-10s  %-20s  %s\n",
			job.ID,
			statusColor(job.Status),
			outputType,
			created,
			filename,
		)
	}

	fmt.Printf("\n%sShowing %d of %d jobs%s\n",
		colorDim, len(resp.Jobs), resp.Pagination.Total, colorReset,
	)

	return nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Local().Format("2006-01-02 15:04")
}
