package deployment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
)

// TelemetryClient sends deployment telemetry over HTTP
type TelemetryClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// NewTelemetryClient creates a new TelemetryClient
func NewTelemetryClient() *TelemetryClient {
	baseURL := os.Getenv("SHOPMON_BASE_URL")
	if baseURL == "" {
		baseURL = "https://shopmon.fos.gg"
	}

	return &TelemetryClient{
		BaseURL:   baseURL,
		AuthToken: os.Getenv("SHOPMON_API_KEY"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendAndParseResponse sends the telemetry payload, uploads output to the presigned URL, and returns the parsed response
func (c *TelemetryClient) SendAndParseResponse(payload *Payload, output string) (map[string]interface{}, error) {
	url := c.BaseURL + "/trpc/cli.createDeployment"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// tRPC returns {"result": {"data": <data>}}
	var trpcResponse map[string]interface{}
	if err := json.Unmarshal(body, &trpcResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	var data map[string]interface{}
	if result, ok := trpcResponse["result"].(map[string]interface{}); ok {
		if d, ok := result["data"].(map[string]interface{}); ok {
			data = d
		}
	}
	if data == nil {
		data = trpcResponse
	}

	if uploadURL, ok := data["upload_url"].(string); ok && uploadURL != "" {
		if err := uploadOutput(c.HTTPClient, uploadURL, output); err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: Failed to upload deployment output: %v\n", err)
		}
	}

	return data, nil
}

// uploadOutput zstd-compresses the output and PUTs it to the presigned S3 URL
func uploadOutput(client *http.Client, uploadURL string, output string) error {
	var buf bytes.Buffer
	encoder, err := zstd.NewWriter(&buf)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	if _, err := encoder.Write([]byte(output)); err != nil {
		return fmt.Errorf("failed to compress output: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to finalize compression: %w", err)
	}

	req, err := http.NewRequest("PUT", uploadURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload output: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// getVersionReference returns the version reference from env var or git commit SHA
func getVersionReference() string {
	if ref := os.Getenv("SHOPMON_DEPLOYMENT_VERSION_REFERENCE"); ref != "" {
		return ref
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// BuildPayload creates a telemetry payload from execution result and composer data
func BuildPayload(result *ExecutionResult, command string, composerData map[string]interface{}) *Payload {
	shopId, _ := strconv.Atoi(os.Getenv("SHOPMON_SHOP_ID"))
	return &Payload{
		ShopId:           shopId,
		Command:          command,
		ReturnCode:       result.ReturnCode,
		StartDate:        result.StartDate,
		EndDate:          result.EndDate,
		ExecutionTime:    result.ExecutionTime,
		Composer:         composerData,
		VersionReference: getVersionReference(),
	}
}
