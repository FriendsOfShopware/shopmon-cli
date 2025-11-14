package deployment

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping test: unable to listen on loopback: %v", err)
	}

	server := &httptest.Server{
		Listener: ln,
		Config:   &http.Server{Handler: handler},
	}

	server.Start()
	t.Cleanup(server.Close)

	return server
}

func TestBuildPayload(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Second)

	result := &ExecutionResult{
		Output:        "command output",
		ReturnCode:    0,
		StartTime:     startTime,
		EndTime:       endTime,
		ExecutionTime: 2.0,
	}

	composerData := map[string]interface{}{
		"php":          ">=8.1",
		"shopware/core": "6.5.0.0",
	}

	payload := BuildPayload(result, "echo test", composerData)

	assert.Equal(t, "echo test", payload.Command)
	assert.Equal(t, "command output", payload.Output)
	assert.Equal(t, 0, payload.ReturnCode)
	assert.Equal(t, startTime.Format(time.RFC3339), payload.StartDate)
	assert.Equal(t, endTime.Format(time.RFC3339), payload.EndDate)
	assert.Equal(t, 2.0, payload.ExecutionTime)
	assert.Equal(t, composerData, payload.Composer)
}

func TestHTTPTelemetryClient_SendAndParseResponse(t *testing.T) {
	t.Run("successful send returns parsed response", func(t *testing.T) {
		var receivedPayload *Payload
		var receivedAuth string
		server := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/cli/deployments", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			receivedAuth = r.Header.Get("Authorization")

			var p Payload
			err := json.NewDecoder(r.Body).Decode(&p)
			require.NoError(t, err)
			receivedPayload = &p

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","deployment_id":"abc123","url":"https://example.test"}`))
		}))
		client := &HTTPTelemetryClient{
			BaseURL:    server.URL,
			AuthToken:  "test-token-123",
			HTTPClient: &http.Client{Timeout: 5 * time.Second},
		}

		payload := &Payload{
			Command:       "test command",
			Output:        "test output",
			ReturnCode:    0,
			StartDate:     time.Now().Format(time.RFC3339),
			EndDate:       time.Now().Add(1 * time.Second).Format(time.RFC3339),
			ExecutionTime: 1.0,
			Composer: map[string]interface{}{
				"php": ">=8.1",
			},
		}

		response, err := client.SendAndParseResponse(payload)
		assert.NoError(t, err)

		assert.NotNil(t, receivedPayload)
		assert.Equal(t, payload.Command, receivedPayload.Command)
		assert.Equal(t, payload.Output, receivedPayload.Output)
		assert.Equal(t, payload.ReturnCode, receivedPayload.ReturnCode)
		assert.Equal(t, "Bearer test-token-123", receivedAuth)
		assert.Equal(t, "abc123", response["deployment_id"])
		assert.Equal(t, "https://example.test", response["url"])
	})

	t.Run("server error", func(t *testing.T) {
		server := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))

		client := &HTTPTelemetryClient{
			BaseURL:    server.URL,
			HTTPClient: &http.Client{Timeout: 5 * time.Second},
		}

		payload := &Payload{
			Command:    "test",
			Output:     "output",
			ReturnCode: 0,
		}

		response, err := client.SendAndParseResponse(payload)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
		assert.Contains(t, err.Error(), "Internal Server Error")
	})

	t.Run("network error", func(t *testing.T) {
		client := &HTTPTelemetryClient{
			BaseURL:    "http://nonexistent.local:12345",
			HTTPClient: &http.Client{Timeout: 1 * time.Second},
		}

		payload := &Payload{
			Command:    "test",
			Output:     "output",
			ReturnCode: 0,
		}

		response, err := client.SendAndParseResponse(payload)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send request")
	})

	t.Run("invalid payload", func(t *testing.T) {
		client := &HTTPTelemetryClient{
			BaseURL:    "http://example.com",
			HTTPClient: &http.Client{},
		}

		payload := &Payload{
			Command: "test",
			Composer: map[string]interface{}{
				"invalid": make(chan int), // channels can't be marshaled to JSON
			},
		}

		response, err := client.SendAndParseResponse(payload)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal payload")
	})

	t.Run("404 not found", func(t *testing.T) {
		server := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))

		client := &HTTPTelemetryClient{
			BaseURL:    server.URL,
			HTTPClient: &http.Client{Timeout: 5 * time.Second},
		}

		payload := &Payload{
			Command:    "test",
			Output:     "output",
			ReturnCode: 0,
		}

		response, err := client.SendAndParseResponse(payload)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 404")
	})

	t.Run("without authorization token", func(t *testing.T) {
		var receivedAuth string
		server := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}))

		client := &HTTPTelemetryClient{
			BaseURL:    server.URL,
			AuthToken:  "", // No token
			HTTPClient: &http.Client{Timeout: 5 * time.Second},
		}

		payload := &Payload{
			Command:    "test",
			Output:     "output",
			ReturnCode: 0,
		}

		response, err := client.SendAndParseResponse(payload)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, receivedAuth, "Should not have Authorization header when no token is set")
	})

	t.Run("authorization failure", func(t *testing.T) {
		server := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorized"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		client := &HTTPTelemetryClient{
			BaseURL:    server.URL,
			AuthToken:  "invalid-token",
			HTTPClient: &http.Client{Timeout: 5 * time.Second},
		}

		payload := &Payload{
			Command:    "test",
			Output:     "output",
			ReturnCode: 0,
		}

		response, err := client.SendAndParseResponse(payload)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 401")
		assert.Contains(t, err.Error(), "Unauthorized")
	})
}

func TestNewHTTPTelemetryClient(t *testing.T) {
	t.Run("default URL and no token", func(t *testing.T) {
		// Ensure env vars are not set
		t.Setenv("SHOPMON_BASE_URL", "")
		t.Setenv("SHOPMON_DEPLOY_TOKEN", "")

		client := NewHTTPTelemetryClient()
		assert.Equal(t, "https://shopmon.fos.gg", client.BaseURL)
		assert.Empty(t, client.AuthToken)
		assert.NotNil(t, client.HTTPClient)
		assert.Equal(t, 30*time.Second, client.HTTPClient.Timeout)
	})

	t.Run("custom URL from environment", func(t *testing.T) {
		t.Setenv("SHOPMON_BASE_URL", "http://custom.example.com")

		client := NewHTTPTelemetryClient()
		assert.Equal(t, "http://custom.example.com", client.BaseURL)
	})

	t.Run("auth token from environment", func(t *testing.T) {
		t.Setenv("SHOPMON_DEPLOY_TOKEN", "my-secret-token")

		client := NewHTTPTelemetryClient()
		assert.Equal(t, "my-secret-token", client.AuthToken)
	})

	t.Run("both URL and token from environment", func(t *testing.T) {
		t.Setenv("SHOPMON_BASE_URL", "http://staging.example.com")
		t.Setenv("SHOPMON_DEPLOY_TOKEN", "staging-token-456")

		client := NewHTTPTelemetryClient()
		assert.Equal(t, "http://staging.example.com", client.BaseURL)
		assert.Equal(t, "staging-token-456", client.AuthToken)
		assert.NotNil(t, client.HTTPClient)
	})
}

func TestPayload_JSONSerialization(t *testing.T) {
	payload := &Payload{
		Command:       "echo test",
		Output:        "test output\nwith newlines",
		ReturnCode:    0,
		StartDate:     "2024-01-01T12:00:00Z",
		EndDate:       "2024-01-01T12:00:01Z",
		ExecutionTime: 1.0,
		Composer: map[string]interface{}{
			"php":           ">=8.1",
			"shopware/core": "6.5.0.0",
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	// Deserialize back
	var decoded Payload
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, payload.Command, decoded.Command)
	assert.Equal(t, payload.Output, decoded.Output)
	assert.Equal(t, payload.ReturnCode, decoded.ReturnCode)
	assert.Equal(t, payload.StartDate, decoded.StartDate)
	assert.Equal(t, payload.EndDate, decoded.EndDate)
	assert.Equal(t, payload.ExecutionTime, decoded.ExecutionTime)
	assert.Equal(t, payload.Composer, decoded.Composer)

	// Verify JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err)

	assert.Contains(t, jsonMap, "command")
	assert.Contains(t, jsonMap, "output")
	assert.Contains(t, jsonMap, "return_code")
	assert.Contains(t, jsonMap, "start_date")
	assert.Contains(t, jsonMap, "end_date")
	assert.Contains(t, jsonMap, "execution_time")
	assert.Contains(t, jsonMap, "composer")
}
