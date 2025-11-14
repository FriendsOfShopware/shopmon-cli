package deployment

import (
	"fmt"
	"os"
)

// Service orchestrates the deployment command execution
type Service struct {
	Executor       CommandExecutor
	ComposerReader ComposerReader
	Telemetry      TelemetryClient
}

// NewService creates a new deployment service with default implementations
func NewService() *Service {
	return &Service{
		Executor:       NewDefaultExecutor(),
		ComposerReader: NewDefaultComposerReader(),
		Telemetry:      NewHTTPTelemetryClient(),
	}
}

// Run executes a deployment command and sends telemetry
func (s *Service) Run(args []string) error {
	// Parse command from args
	command, err := ParseCommand(args)
	if err != nil {
		return fmt.Errorf("usage: deploy -- <command>\n\nExample:\n  shopmon-cli deploy -- php artisan migrate")
	}

	// Execute the command
	result, err := s.Executor.Execute(command)
	if err != nil && result == nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Print the output to stdout (transparent execution)
	fmt.Print(result.Output)
	os.Stdout.Sync()

	// Read composer.json if it exists
	composerData, err := s.ComposerReader.ReadComposerData("composer.json")
	if err != nil {
		// Log warning but continue
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to read composer.json: %v\n", err)
		composerData = make(map[string]interface{})
	}

	// Build telemetry payload
	payload := BuildPayload(result, command, composerData)

	// Send to monitoring service and get response
	response, err := s.Telemetry.SendAndParseResponse(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: Failed to send telemetry: %v\n", err)
	} else if response != nil {
		// Display deployment URL if available
		if url, ok := response["url"].(string); ok && url != "" {
			fmt.Fprintf(os.Stderr, "\nDeployment URL: %s\n", url)
		}
		// Display deployment ID if available
		if deploymentID, ok := response["deployment_id"].(string); ok && deploymentID != "" {
			fmt.Fprintf(os.Stderr, "Deployment ID: %s\n", deploymentID)
		}
	}

	// Exit with the same code as the executed command
	if result.ReturnCode != 0 {
		os.Exit(result.ReturnCode)
	}

	return nil
}