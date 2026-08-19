package redis

import (
	"encoding/json"
	"testing"
)

func TestGetValidation(t *testing.T) {
	_, err := Get("", "", 0, "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSetValidation(t *testing.T) {
	_, err := Set("", "", "", 0, "", 0)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestDelValidation(t *testing.T) {
	_, err := Del(nil, "", 0, "")
	if err == nil {
		t.Fatal("expected error for empty keys")
	}
}

func TestPingNotRunning(t *testing.T) {
	// Redis is likely not running on CI
	res, err := Ping("127.0.0.1", 6379, "")
	if err != nil {
		t.Fatal(err)
	}
	// Just verify no crash; Up may be false
	_ = res
}

func TestGetNotFound(t *testing.T) {
	// Without Redis running, this should error
	_, err := Get("nonexistent_key_xyz", "127.0.0.1", 6379, "")
	// Error is expected if redis-cli not available
	_ = err
}

func TestDefaultOpts(t *testing.T) {
	o := defaultOpts()
	if o.host != "127.0.0.1" || o.port != 6379 {
		t.Fatalf("unexpected defaults: %+v", o)
	}
}

func TestParseOpts(t *testing.T) {
	o := parseOpts("localhost", 6380, "secret")
	if o.host != "localhost" || o.port != 6380 || o.auth != "secret" {
		t.Fatalf("unexpected: %+v", o)
	}
}

func TestParseOptsDefaults(t *testing.T) {
	o := parseOpts("", 0, "")
	if o.host != "127.0.0.1" || o.port != 6379 {
		t.Fatalf("expected defaults: %+v", o)
	}
}

func TestPingResultJSON(t *testing.T) {
	r := PingResult{Up: true, Response: "PONG"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded PingResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Up || decoded.Response != "PONG" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestGetResultJSON(t *testing.T) {
	r := GetResult{Key: "test", Value: "hello", Found: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GetResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Found || decoded.Value != "hello" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestSetResultJSON(t *testing.T) {
	r := SetResult{Key: "test", Success: true, Duration: 5}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SetResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Success {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestDelResultJSON(t *testing.T) {
	r := DelResult{Keys: []string{"a", "b"}, Deleted: 2, Success: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DelResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Deleted != 2 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestListResultJSON(t *testing.T) {
	r := ListResult{Pattern: "*", Keys: []string{"a"}, Count: 1}
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

func TestInfoResultJSON(t *testing.T) {
	r := InfoResult{Version: "7.0.0", Uptime: 3600}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded InfoResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != "7.0.0" {
		t.Fatalf("unexpected: %+v", decoded)
	}
}
