package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StemSplit/stemsplit-cli/internal/client"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var separateCmd = &cobra.Command{
	Use:   "separate <file>",
	Short: "Separate audio stems from a file",
	Long: `Upload an audio file and separate it into individual stems.

Examples:
  stemsplit separate song.mp3
  stemsplit separate song.mp3 --stems vocals --output ./out/
  stemsplit separate song.mp3 --stems vocals,drums,bass,other --format wav`,
	Args: cobra.ExactArgs(1),
	RunE: runSeparate,
}

var (
	separateStems        string
	separateModel        string
	separateOutput       string
	separateFormat       string
	separateWait         bool
	separatePollInterval int
)

func init() {
	separateCmd.Flags().StringVarP(&separateStems, "stems", "s", "vocals,drums,bass,other", "stems to extract: vocals, instrumental, both, vocals/drums/bass/other (4-stem), or add piano/guitar (6-stem)")
	separateCmd.Flags().StringVarP(&separateModel, "model", "m", "htdemucs_ft", "model/quality: htdemucs_ft (best), htdemucs (balanced), fast")
	separateCmd.Flags().StringVarP(&separateOutput, "output", "o", ".", "output directory for downloaded stems")
	separateCmd.Flags().StringVar(&separateFormat, "format", "mp3", "output format: mp3, wav, flac")
	separateCmd.Flags().BoolVarP(&separateWait, "wait", "w", true, "wait for completion and download stems")
	separateCmd.Flags().IntVar(&separatePollInterval, "poll-interval", 3, "seconds between status checks")
}

func runSeparate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	apiKey, err := resolveAPIKey()
	if err != nil {
		return err
	}

	c := client.New(apiKey)

	outputType := stemsToOutputType(separateStems)
	quality := modelToQuality(separateModel)
	format := strings.ToUpper(separateFormat)

	filename := filepath.Base(filePath)

	// Step 1: Get presigned upload URL
	spin := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	spin.Suffix = "  Getting upload URL..."
	spin.Start()

	uploadResp, err := c.GetUploadURL(filename, "")
	if err != nil {
		spin.Stop()
		return fmt.Errorf("failed to get upload URL: %w", err)
	}
	spin.Stop()

	// Step 2: Upload the file
	spin = spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	spin.Suffix = fmt.Sprintf("  Uploading %s...", filename)
	spin.Start()

	if err := c.UploadFile(uploadResp.UploadURL, filePath, uploadResp.ContentType); err != nil {
		spin.Stop()
		return fmt.Errorf("upload failed: %w", err)
	}
	spin.Stop()
	fmt.Printf("%s✓%s Uploaded %s\n", colorGreen, colorReset, filename)

	// Step 3: Create the job
	jobReq := &client.CreateJobRequest{
		UploadKey:    uploadResp.UploadKey,
		FileName:     filename,
		OutputType:   outputType,
		Quality:      quality,
		OutputFormat: format,
	}

	job, err := c.CreateJob(jobReq)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	fmt.Printf("%s✓%s Job created: %s%s%s\n", colorGreen, colorReset, colorBold, job.ID, colorReset)
	fmt.Printf("  Output type: %s  Quality: %s  Format: %s\n",
		strings.ToLower(outputType),
		strings.ToLower(quality),
		strings.ToLower(format),
	)

	if !separateWait {
		fmt.Printf("\nJob submitted. Check status with:\n  stemsplit jobs\n")
		return nil
	}

	// Step 4: Poll until done
	fmt.Println()
	spin = spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	spin.Suffix = "  Processing... (this may take a minute)"
	spin.Start()

	for {
		time.Sleep(time.Duration(separatePollInterval) * time.Second)

		job, err = c.GetJob(job.ID)
		if err != nil {
			spin.Stop()
			return fmt.Errorf("failed to check job status: %w", err)
		}

		switch job.Status {
		case "COMPLETED":
			spin.Stop()
			goto download
		case "FAILED":
			spin.Stop()
			msg := "processing failed"
			if job.ErrorMessage != nil {
				msg = *job.ErrorMessage
			}
			return fmt.Errorf("job failed: %s", msg)
		default:
			if job.Progress > 0 {
				spin.Suffix = fmt.Sprintf("  Processing... %d%%", job.Progress)
			}
		}
	}

download:
	fmt.Printf("%s✓%s Processing complete!\n\n", colorGreen, colorReset)

	// Step 5: Download stems
	if err := os.MkdirAll(separateOutput, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	ext := "." + strings.ToLower(separateFormat)
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

	downloadCount := 0
	for stemName, output := range job.Outputs {
		if output == nil || output.URL == "" {
			continue
		}

		destFile := filepath.Join(separateOutput, fmt.Sprintf("%s_%s%s", baseName, stemName, ext))

		spin = spinner.New(spinner.CharSets[14], 80*time.Millisecond)
		spin.Suffix = fmt.Sprintf("  Downloading %s...", stemName)
		spin.Start()

		if err := c.DownloadFile(output.URL, destFile); err != nil {
			spin.Stop()
			fmt.Printf("  %s✗%s Failed to download %s: %v\n", colorRed, colorReset, stemName, err)
			continue
		}
		spin.Stop()
		fmt.Printf("  %s✓%s %s → %s\n", colorGreen, colorReset, stemName, destFile)
		downloadCount++
	}

	fmt.Printf("\n%s%d stem(s) saved to %s%s\n", colorBold, downloadCount, separateOutput, colorReset)
	return nil
}

// stemsToOutputType maps comma-separated stem names to an API outputType.
func stemsToOutputType(stems string) string {
	parts := splitTrim(stems)
	stemSet := make(map[string]bool, len(parts))
	for _, p := range parts {
		stemSet[strings.ToLower(p)] = true
	}

	// Single-stem shortcuts
	if len(parts) == 1 {
		switch parts[0] {
		case "vocals":
			return "VOCALS"
		case "instrumental", "no-vocals", "novocals":
			return "INSTRUMENTAL"
		case "both":
			return "BOTH"
		case "four", "four-stems", "4stems":
			return "FOUR_STEMS"
		case "six", "six-stems", "6stems":
			return "SIX_STEMS"
		}
	}

	// Two-stem vocals+instrumental
	if len(parts) == 2 && stemSet["vocals"] && stemSet["instrumental"] {
		return "BOTH"
	}

	// Six or more → SIX_STEMS
	if len(parts) >= 6 {
		return "SIX_STEMS"
	}

	// Default: four stems
	return "FOUR_STEMS"
}

// modelToQuality maps a model name to an API quality level.
func modelToQuality(model string) string {
	switch strings.ToLower(model) {
	case "htdemucs_ft", "htdemucs-ft", "best":
		return "BEST"
	case "htdemucs", "balanced":
		return "BALANCED"
	case "spleeter", "fast":
		return "FAST"
	default:
		return "BEST"
	}
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
