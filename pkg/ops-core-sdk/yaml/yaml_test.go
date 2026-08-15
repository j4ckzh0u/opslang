package opsyaml

import (
	"strings"
	"testing"
)

func TestEncode_Struct(t *testing.T) {
	type Person struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	result, err := Encode(Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Size == 0 {
		t.Fatal("expected non-zero size")
	}
	if !strings.Contains(result.YAML, "Alice") {
		t.Fatalf("YAML missing name: %s", result.YAML)
	}
	if !strings.Contains(result.YAML, "30") {
		t.Fatalf("YAML missing age: %s", result.YAML)
	}
}

func TestEncode_Map(t *testing.T) {
	data := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}
	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.YAML, "key") || !strings.Contains(result.YAML, "value") {
		t.Fatalf("YAML missing expected values: %s", result.YAML)
	}
}

func TestEncode_Slice(t *testing.T) {
	data := []string{"a", "b", "c"}
	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.YAML, "a") || !strings.Contains(result.YAML, "c") {
		t.Fatalf("YAML missing expected values: %s", result.YAML)
	}
}

func TestEncode_Nested(t *testing.T) {
	data := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	}
	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.YAML, "outer") || !strings.Contains(result.YAML, "inner") {
		t.Fatalf("YAML missing nested keys: %s", result.YAML)
	}
}

func TestEncode_Unsupported(t *testing.T) {
	_, err := Encode(make(chan int))
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestDecode_ValidMapping(t *testing.T) {
	input := "name: Alice\nage: 30\n"
	result, err := Decode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result.Data)
	}
	if m["name"] != "Alice" {
		t.Fatalf("expected Alice, got %v", m["name"])
	}
	// YAML integers decode as int
	if m["age"] != 30 {
		t.Fatalf("expected 30, got %v", m["age"])
	}
}

func TestDecode_Sequence(t *testing.T) {
	input := "- 1\n- 2\n- 3\n"
	result, err := Decode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := result.Data.([]interface{})
	if !ok {
		t.Fatalf("expected slice, got %T", result.Data)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
}

func TestDecode_Scalar(t *testing.T) {
	result, err := Decode("hello\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != "hello" {
		t.Fatalf("expected hello, got %v (%T)", result.Data, result.Data)
	}
}

func TestDecode_Number(t *testing.T) {
	result, err := Decode("3.14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != 3.14 {
		t.Fatalf("expected 3.14, got %v", result.Data)
	}
}

func TestDecode_Boolean(t *testing.T) {
	result, err := Decode("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != true {
		t.Fatalf("expected true, got %v", result.Data)
	}
}

func TestDecode_Null(t *testing.T) {
	result, err := Decode("null")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != nil {
		t.Fatalf("expected nil, got %v", result.Data)
	}
}

func TestDecode_InvalidYAML(t *testing.T) {
	_, err := Decode(":\n  - :\n  invalid: [unterminated")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestEncodeDecode_Roundtrip(t *testing.T) {
	original := map[string]interface{}{
		"key":   "value",
		"count": 42,
		"nested": map[string]interface{}{
			"flag": true,
		},
	}
	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := Decode(encoded.YAML)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	m, ok := decoded.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", decoded.Data)
	}
	if m["key"] != "value" {
		t.Fatalf("roundtrip mismatch for key: %v", m["key"])
	}
}
