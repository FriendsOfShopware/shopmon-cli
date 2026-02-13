package deployment

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
)

// Execute runs a command and captures its output
func Execute(command string) (*ExecutionResult, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, exec.ErrNotFound
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	startTime := time.Now()
	err := cmd.Run()
	endTime := time.Now()

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
		StartDate:     startTime.Format(time.RFC3339),
		EndDate:       endTime.Format(time.RFC3339),
		ExecutionTime: endTime.Sub(startTime).Seconds(),
	}, nil
}

// ParseCommand extracts the command string from args after the "--" separator
func ParseCommand(args []string) (string, error) {
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

	return strings.Join(args[dashIndex+1:], " "), nil
}
