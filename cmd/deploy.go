package cmd

import (
	"fmt"
	"os"

	"github.com/friendsofshopware/shopmon-cli/internal/deployment"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:                "deploy -- <command>",
	Short:              "Execute and monitor deployment commands",
	Long:               `Execute deployment commands, capture output, and send telemetry to the monitoring service.`,
	Example:            "  shopmon-cli deploy -- php artisan migrate\n  shopmon-cli deploy -- composer install",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if token := os.Getenv("SHOPMON_API_KEY"); token == "" {
			return fmt.Errorf("SHOPMON_API_KEY environment variable must be set to use this command")
		}

		return deployment.Run(args)
	},
}
