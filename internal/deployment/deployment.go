package deployment

import (
	"fmt"
	"os"
)

// Run executes a deployment command and sends telemetry
func Run(args []string) error {
	command, err := ParseCommand(args)
	if err != nil {
		return fmt.Errorf("usage: deploy -- <command>\n\nExample:\n  shopmon-cli deploy -- php artisan migrate")
	}

	result, err := Execute(command)
	if err != nil && result == nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Print(result.Output)

	composerData, err := ReadComposerData("composer.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to read composer.json: %v\n", err)
		composerData = make(map[string]interface{})
	}

	payload := BuildPayload(result, command, composerData)

	telemetry := NewTelemetryClient()
	response, err := telemetry.SendAndParseResponse(payload, result.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to send telemetry: %v\n", err)
	} else if response != nil {
		if url, ok := response["url"].(string); ok && url != "" {
			fmt.Fprintf(os.Stderr, "\nDeployment URL: %s\n", url)
		}
		if deploymentID, ok := response["deployment_id"].(string); ok && deploymentID != "" {
			fmt.Fprintf(os.Stderr, "Deployment ID: %s\n", deploymentID)
		}
	}

	if result.ReturnCode != 0 {
		os.Exit(result.ReturnCode)
	}

	return nil
}
