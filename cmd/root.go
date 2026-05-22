package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "stemsplit",
	Short: "AI-powered audio stem separation from the command line",
	Long: `stemsplit is the official CLI for StemSplit.

Separate vocals, drums, bass, and more from any audio file using
the StemSplit API. Get your API key at https://stemsplit.io.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute(version string) {
	Version = version
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, colorRed+"Error:"+colorReset+" "+err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(separateCmd)
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(configCmd)
}

// --- Config helpers ---

type Config struct {
	APIKey string `json:"api_key"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "stemsplit")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func loadConfig() *Config {
	cfg := &Config{}
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, cfg)
	return cfg
}

func saveConfig(cfg *Config) error {
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

func resolveAPIKey() (string, error) {
	if key := os.Getenv("STEMSPLIT_API_KEY"); key != "" {
		return key, nil
	}
	cfg := loadConfig()
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	return "", fmt.Errorf(
		"no API key found\n\nSet it with:\n  export STEMSPLIT_API_KEY=sk_live_xxx\n  stemsplit config set api_key sk_live_xxx\n\nGet your key at https://stemsplit.io",
	)
}

// --- ANSI color constants ---

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func statusColor(status string) string {
	switch status {
	case "COMPLETED":
		return colorGreen + status + colorReset
	case "FAILED":
		return colorRed + status + colorReset
	case "PROCESSING":
		return colorYellow + status + colorReset
	default:
		return colorDim + status + colorReset
	}
}
