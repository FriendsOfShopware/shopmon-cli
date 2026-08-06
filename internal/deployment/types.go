package deployment

// Payload represents the deployment telemetry data sent to the monitoring service
type Payload struct {
	EnvironmentId    int     `json:"environment_id"`
	Name             string  `json:"name,omitempty"`
	Command          string  `json:"command"`
	ReturnCode       int     `json:"return_code"`
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date"`
	ExecutionTime    float64 `json:"execution_time"`
	Composer         *string `json:"composer,omitempty"`
	VersionReference string  `json:"reference,omitempty"`
}

// Response represents the API response for a created deployment
type Response struct {
	Success      bool   `json:"success"`
	Name         string `json:"name"`
	DeploymentId int    `json:"deployment_id"`
	URL          string `json:"url"`
	UploadURL    string `json:"upload_url"`
}

// ExecutionResult holds the result of a command execution
type ExecutionResult struct {
	Output        string
	ReturnCode    int
	StartDate     string
	EndDate       string
	ExecutionTime float64
}
