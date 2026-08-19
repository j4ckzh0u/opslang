package consul

import (
	"encoding/json"
	"testing"
)

func TestKVGetValidation(t *testing.T) {
	r := KVGet("", "")
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestKVPutValidation(t *testing.T) {
	r := KVPut("", "value", "")
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestKVDeleteValidation(t *testing.T) {
	r := KVDelete("", "")
	if r.Error == "" {
		t.Error("empty key should return error")
	}
}

func TestKVListValidation(t *testing.T) {
	_, err := KVList("", "")
	if err == nil {
		t.Error("empty prefix should return error")
	}
}

func TestServiceRegisterValidation(t *testing.T) {
	r := ServiceRegister("", "", "", "", "")
	if r.Error == "" {
		t.Error("empty name should return error")
	}
}

func TestServiceDeregisterValidation(t *testing.T) {
	r := ServiceDeregister("", "")
	if r.Error == "" {
		t.Error("empty id should return error")
	}
}

func TestHealthCheckValidation(t *testing.T) {
	r := HealthCheck("", "")
	if r.Error == "" {
		t.Error("empty service should return error")
	}
}

func TestJSONSerialization(t *testing.T) {
	r := KVResult{Key: "test/key", Value: "hello", Success: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out KVResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.Key != "test/key" || out.Value != "hello" {
		t.Errorf("roundtrip failed: %+v", out)
	}
}

func TestMembersResultJSON(t *testing.T) {
	r := MembersResult{Members: []MemberInfo{{Name: "node1", Addr: "10.0.0.1"}}, Count: 1}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out MembersResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out.Count != 1 {
		t.Errorf("expected count 1, got %d", out.Count)
	}
}

func TestInfoResultJSON(t *testing.T) {
	r := InfoResult{Datacenter: "dc1", NodeName: "node1", Ready: true}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out InfoResult
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !out.Ready {
		t.Error("expected ready=true")
	}
}
