package deployment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TelemetryClient interface for sending telemetry data (allows mocking in tests)
type TelemetryClient interface {
	SendAndParseResponse(payload *Payload) (map[string]interface{}, error)
}

// HTTPTelemetryClient implements TelemetryClient using HTTP
type HTTPTelemetryClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// NewHTTPTelemetryClient creates a new HTTPTelemetryClient
func NewHTTPTelemetryClient() *HTTPTelemetryClient {
	baseURL := os.Getenv("SHOPMON_BASE_URL")
	if baseURL == "" {
		baseURL = "https://shopmon.fos.gg"
	}

	authToken := os.Getenv("SHOPMON_DEPLOY_TOKEN")

	return &HTTPTelemetryClient{
		BaseURL:   baseURL,
		AuthToken: authToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}


// SendAndParseResponse sends the telemetry payload and returns the parsed response
func (c *HTTPTelemetryClient) SendAndParseResponse(payload *Payload) (map[string]interface{}, error) {
	url := c.BaseURL + "/api/cli/deployments"

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authorization header if token is present
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return response, nil
}

// BuildPayload creates a telemetry payload from execution result and composer data
func BuildPayload(result *ExecutionResult, command string, composerData map[string]interface{}) *Payload {
	return &Payload{
		Command:       command,
		Output:        result.Output,
		ReturnCode:    result.ReturnCode,
		StartDate:     result.StartTime.Format(time.RFC3339),
		EndDate:       result.EndTime.Format(time.RFC3339),
		ExecutionTime: result.ExecutionTime,
		Composer:      composerData,
	}
}
