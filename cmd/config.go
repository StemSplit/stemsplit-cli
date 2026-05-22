package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Get or set CLI configuration values stored in ~/.config/stemsplit/config.json.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Example: `  stemsplit config set api_key sk_live_xxxxxxxxxxxxx`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Example: `  stemsplit config get api_key`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])
	value := args[1]

	if !isValidConfigKey(key) {
		return fmt.Errorf("unknown config key: %q (valid keys: api_key)", key)
	}

	cfg := loadConfig()
	switch key {
	case "api_key":
		cfg.APIKey = value
	}

	if err := saveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("%s✓%s Set %s → %s\n", colorGreen, colorReset, key, maskSecret(value))
	fmt.Printf("  Config saved to %s\n", configPath())
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(args[0])

	if !isValidConfigKey(key) {
		return fmt.Errorf("unknown config key: %q (valid keys: api_key)", key)
	}

	cfg := loadConfig()
	switch key {
	case "api_key":
		if cfg.APIKey == "" {
			fmt.Printf("%s (not set)\n", key)
		} else {
			fmt.Printf("%s = %s\n", key, maskSecret(cfg.APIKey))
		}
	}
	return nil
}

func isValidConfigKey(key string) bool {
	return key == "api_key"
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:6] + strings.Repeat("*", len(s)-6)
}
