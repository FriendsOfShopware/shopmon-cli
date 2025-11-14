package cmd

import (
	"fmt"
	"os"

	"github.com/friendsofshopware/shopmon-cli/internal/deployment"
	"github.com/spf13/cobra"
)

var newDeploymentService = deployment.NewService

var runDeploymentService = func(service *deployment.Service, args []string) error {
	return service.Run(args)
}

var deployCmd = &cobra.Command{
	Use:                "deploy -- <command>",
	Short:              "Execute and monitor deployment commands",
	Long:               `Execute deployment commands, capture output, and send telemetry to the monitoring service.`,
	Example:            "  shopmon-cli deploy -- php artisan migrate\n  shopmon-cli deploy -- composer install",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if this is a help request
		for _, arg := range args {
			if arg == "-h" || arg == "--help" {
				return cmd.Help()
			}
		}

		if token := os.Getenv("SHOPMON_DEPLOY_TOKEN"); token == "" {
			return fmt.Errorf("SHOPMON_DEPLOY_TOKEN environment variable must be set to use this command")
		}

		// Create deployment service and run
		service := newDeploymentService()
		return runDeploymentService(service, args)
	},
}

func init() {
	// No flags needed since we're using everything after "--"
}
