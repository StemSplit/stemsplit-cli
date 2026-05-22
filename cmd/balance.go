package cmd

import (
	"fmt"

	"github.com/StemSplit/stemsplit-cli/internal/client"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show your credit balance",
	Long:  `Show the current credit balance on your StemSplit account.`,
	RunE:  runBalance,
}

func runBalance(cmd *cobra.Command, args []string) error {
	apiKey, err := resolveAPIKey()
	if err != nil {
		return err
	}

	c := client.New(apiKey)
	resp, err := c.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	minutes := resp.BalanceSeconds / 60
	seconds := resp.BalanceSeconds % 60

	color := colorGreen
	if minutes < 2 {
		color = colorYellow
	}
	if resp.BalanceSeconds == 0 {
		color = colorRed
	}

	fmt.Printf("\n%s%sBalance:%s %s%dm %02ds%s (%d seconds)\n",
		colorBold, colorCyan, colorReset,
		color, minutes, seconds, colorReset,
		resp.BalanceSeconds,
	)

	if minutes < 2 && resp.BalanceSeconds > 0 {
		fmt.Printf("\n%s⚠  Low balance — top up at https://stemsplit.io%s\n", colorYellow, colorReset)
	} else if resp.BalanceSeconds == 0 {
		fmt.Printf("\n%s✗  No credits remaining — top up at https://stemsplit.io%s\n", colorRed, colorReset)
	}

	fmt.Println()
	return nil
}
