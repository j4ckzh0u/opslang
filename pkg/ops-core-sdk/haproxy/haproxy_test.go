package haproxy

import (
	"encoding/json"
	"testing"
)

func TestEnableBackendValidation(t *testing.T) {
	_, err := EnableBackend("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestDisableBackendValidation(t *testing.T) {
	_, err := DisableBackend("", "", "")
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestGetStatusNotRunning(t *testing.T) {
	// On CI (macOS), haproxy is likely not installed
	res, err := GetStatus()
	if err != nil {
		t.Fatal(err)
	}
	// Just verify it returns without error; Up may be false
	if res.Up && res.PID == 0 {
		t.Fatal("if Up is true, PID should be > 0")
	}
}

func TestStatusResultJSON(t *testing.T) {
	r := StatusResult{Up: true, PID: 1234}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded StatusResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Up || decoded.PID != 1234 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Backend: "web", Success: true, Changed: true, Duration: 10}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ActionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success || decoded.Backend != "web" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestConfigResultJSON(t *testing.T) {
	r := ConfigResult{Valid: false, Errors: []string{"bad config"}}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ConfigResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Valid || len(decoded.Errors) != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestListResultJSON(t *testing.T) {
	r := ListResult{Backends: []BackendInfo{{Name: "web", Status: "UP", Type: "backend"}}, Count: 1}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Count != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestAtoi(t *testing.T) {
	if atoi("1234") != 1234 {
		t.Fatal("expected 1234")
	}
	if atoi("0") != 0 {
		t.Fatal("expected 0")
	}
	if atoi("") != 0 {
		t.Fatal("expected 0 for empty")
	}
}
