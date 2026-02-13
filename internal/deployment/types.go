package deployment

// Payload represents the deployment telemetry data sent to the monitoring service
type Payload struct {
	ShopId           int                    `json:"shop_id"`
	Command          string                 `json:"command"`
	ReturnCode       int                    `json:"return_code"`
	StartDate        string                 `json:"start_date"`
	EndDate          string                 `json:"end_date"`
	ExecutionTime    float64                `json:"execution_time"`
	Composer         map[string]interface{} `json:"composer"`
	VersionReference string                 `json:"reference,omitempty"`
}

// ExecutionResult holds the result of a command execution
type ExecutionResult struct {
	Output        string
	ReturnCode    int
	StartDate     string
	EndDate       string
	ExecutionTime float64
}
