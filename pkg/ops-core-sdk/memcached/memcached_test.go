package memcached

import (
	"encoding/json"
	"testing"
)

func TestGetValidation(t *testing.T) {
	r := Get("", "", 0)
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestSetValidation(t *testing.T) {
	r := Set("", "value", "", 0, 0)
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestDeleteValidation(t *testing.T) {
	r := Delete("", "", 0)
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestGetResultJSON(t *testing.T) {
	r := GetResult{Key: "test", Value: "hello", Found: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out GetResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !out.Found || out.Value != "hello" {
		t.Errorf("roundtrip failed: %+v", out)
	}
}

func TestSetResultJSON(t *testing.T) {
	r := SetResult{Key: "test", Success: true, Stored: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out SetResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !out.Stored {
		t.Error("expected stored=true")
	}
}

func TestStatsResultJSON(t *testing.T) {
	r := StatsResult{Version: "1.6.0", CurrItems: 100, Uptime: 3600}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out StatsResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.CurrItems != 100 {
		t.Errorf("expected 100 items, got %d", out.CurrItems)
	}
}

func TestVersionResultJSON(t *testing.T) {
	r := VersionResult{Version: "1.6.22"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out VersionResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.Version != "1.6.22" {
		t.Errorf("expected 1.6.22, got %s", out.Version)
	}
}

func TestDialDefaults(t *testing.T) {
	// Test that empty host defaults to 127.0.0.1:11211
	// This will fail to connect but shouldn't panic
	conn, err := dial("", 0)
	if err == nil && conn != nil {
		conn.Close()
	}
}
