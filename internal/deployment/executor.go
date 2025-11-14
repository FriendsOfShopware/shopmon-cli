package deployment

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor interface for executing commands (allows mocking in tests)
type CommandExecutor interface {
	Execute(command string) (*ExecutionResult, error)
}

// DefaultExecutor implements CommandExecutor using os/exec
type DefaultExecutor struct{}

// NewDefaultExecutor creates a new DefaultExecutor
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{}
}

// Execute runs a command and captures its output
func (e *DefaultExecutor) Execute(command string) (*ExecutionResult, error) {
	startTime := time.Now()

	// Parse command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, exec.ErrNotFound
	}

	// Create command
	var cmd *exec.Cmd
	if len(parts) == 1 {
		cmd = exec.Command(parts[0])
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	// Capture output
	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	// Run the command
	err := cmd.Run()

	endTime := time.Now()
	executionTime := endTime.Sub(startTime).Seconds()

	// Get return code
	returnCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			returnCode = exitError.ExitCode()
		} else {
			returnCode = 1
		}
	}

	return &ExecutionResult{
		Output:        outputBuffer.String(),
		ReturnCode:    returnCode,
		StartTime:     startTime,
		EndTime:       endTime,
		ExecutionTime: executionTime,
	}, nil
}

// ParseCommand splits a command string respecting quotes and spaces
func ParseCommand(args []string) (string, error) {
	// Find the "--" separator
	dashIndex := -1
	for i, arg := range args {
		if arg == "--" {
			dashIndex = i
			break
		}
	}

	if dashIndex == -1 || dashIndex == len(args)-1 {
		return "", exec.ErrNotFound
	}

	// Join all arguments after "--" as the command to execute
	commandParts := args[dashIndex+1:]
	return strings.Join(commandParts, " "), nil
}