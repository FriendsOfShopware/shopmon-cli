package deployment

import "time"

// Payload represents the deployment telemetry data sent to the monitoring service
type Payload struct {
	Command       string                 `json:"command"`
	Output        string                 `json:"output"`
	ReturnCode    int                    `json:"return_code"`
	StartDate     string                 `json:"start_date"`
	EndDate       string                 `json:"end_date"`
	ExecutionTime float64                `json:"execution_time"`
	Composer      map[string]interface{} `json:"composer"`
}

// ComposerJSON represents the structure of a composer.json file
type ComposerJSON struct {
	Require map[string]string `json:"require"`
}

// ExecutionResult holds the result of a command execution
type ExecutionResult struct {
	Output        string
	ReturnCode    int
	StartTime     time.Time
	EndTime       time.Time
	ExecutionTime float64
}