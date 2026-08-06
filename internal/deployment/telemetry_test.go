package deployment

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// Mirrors api.CreateCliDeploymentRequest from shopmon/api/internal/api/server.gen.go
type apiRequest struct {
	Command       string    `json:"command"`
	Composer      *string   `json:"composer,omitempty"`
	EndDate       time.Time `json:"end_date"`
	EnvironmentId int       `json:"environment_id"`
	ExecutionTime float32   `json:"execution_time"`
	Name          *string   `json:"name,omitempty"`
	Reference     *string   `json:"reference,omitempty"`
	ReturnCode    int       `json:"return_code"`
	StartDate     time.Time `json:"start_date"`
}

// Mirrors api.CreateCliDeploymentResponse
type apiResponse struct {
	DeploymentId int    `json:"deployment_id"`
	Name         string `json:"name"`
	Success      bool   `json:"success"`
	UploadUrl    string `json:"upload_url"`
	Url          string `json:"url"`
}

func TestPayloadDecodesIntoAPIRequest(t *testing.T) {
	t.Setenv("SHOPMON_ENVIRONMENT_ID", "42")
	t.Setenv("SHOPMON_DEPLOYMENT_VERSION_REFERENCE", "abc123")

	p := BuildPayload(&ExecutionResult{
		ReturnCode:    0,
		StartDate:     "2026-08-06T10:00:00Z",
		EndDate:       "2026-08-06T10:01:00Z",
		ExecutionTime: 60.0,
	}, "bin/console deploy:run", map[string]interface{}{"shopware/core": "6.6.0"})

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("CLI wire payload: %s", raw)

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var req apiRequest
	if err := dec.Decode(&req); err != nil {
		t.Fatalf("API cannot decode CLI payload: %v", err)
	}

	if req.EnvironmentId != 42 {
		t.Errorf("environment_id = %d, want 42", req.EnvironmentId)
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		t.Errorf("dates did not parse: %v %v", req.StartDate, req.EndDate)
	}
	if req.ExecutionTime != 60.0 {
		t.Errorf("execution_time = %v", req.ExecutionTime)
	}
	if req.Composer == nil {
		t.Fatal("composer is nil")
	}
	var composer map[string]string
	if err := json.Unmarshal([]byte(*req.Composer), &composer); err != nil {
		t.Fatalf("composer not valid JSON for jsonb column: %v", err)
	}
	if composer["shopware/core"] != "6.6.0" {
		t.Errorf("composer = %v", composer)
	}
	if req.Reference == nil || *req.Reference != "abc123" {
		t.Errorf("reference = %v", req.Reference)
	}
	t.Logf("decoded OK: env=%d composer=%v ref=%s", req.EnvironmentId, composer, *req.Reference)
}

func TestDeploymentNameForwarded(t *testing.T) {
	t.Setenv("SHOPMON_ENVIRONMENT_ID", "1")
	t.Setenv("SHOPMON_DEPLOYMENT_NAME", "Production Deploy")

	raw, _ := json.Marshal(BuildPayload(&ExecutionResult{
		StartDate: "2026-08-06T10:00:00Z", EndDate: "2026-08-06T10:00:01Z",
	}, "cmd", nil))

	var req apiRequest
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		t.Fatalf("API cannot decode CLI payload: %v", err)
	}
	if req.Name == nil || *req.Name != "Production Deploy" {
		t.Errorf("name = %v, want %q", req.Name, "Production Deploy")
	}
}

func TestDeploymentNameOmittedWhenUnset(t *testing.T) {
	t.Setenv("SHOPMON_ENVIRONMENT_ID", "1")
	t.Setenv("SHOPMON_DEPLOYMENT_NAME", "")

	raw, _ := json.Marshal(BuildPayload(&ExecutionResult{
		StartDate: "2026-08-06T10:00:00Z", EndDate: "2026-08-06T10:00:01Z",
	}, "cmd", nil))

	// Omitting the field lets the server generate the name.
	if bytes.Contains(raw, []byte(`"name"`)) {
		t.Errorf("unset name should be omitted, got %s", raw)
	}
}

func TestLegacyShopIDFallback(t *testing.T) {
	t.Setenv("SHOPMON_ENVIRONMENT_ID", "")
	t.Setenv("SHOPMON_SHOP_ID", "7")
	if got := EnvironmentID(); got != 7 {
		t.Errorf("EnvironmentID() = %d, want 7", got)
	}
}

func TestEmptyComposerOmitted(t *testing.T) {
	t.Setenv("SHOPMON_ENVIRONMENT_ID", "1")
	t.Setenv("SHOPMON_DEPLOYMENT_VERSION_REFERENCE", "x")
	p := BuildPayload(&ExecutionResult{StartDate: "2026-08-06T10:00:00Z", EndDate: "2026-08-06T10:00:01Z"}, "cmd", map[string]interface{}{})
	raw, _ := json.Marshal(p)
	if bytes.Contains(raw, []byte("composer")) {
		t.Errorf("empty composer should be omitted, got %s", raw)
	}
}

func TestResponseParsesAPIShape(t *testing.T) {
	raw, _ := json.Marshal(apiResponse{
		Success: true, Name: "Prod", DeploymentId: 7,
		Url: "https://shopmon.fos.gg/environments/42/deployments/7", UploadUrl: "https://s3/up",
	})
	var got Response
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Success || got.Name != "Prod" || got.DeploymentId != 7 ||
		got.URL == "" || got.UploadURL == "" {
		t.Fatalf("CLI response mismatch: %+v", got)
	}
	t.Logf("response OK: %+v", got)
}
