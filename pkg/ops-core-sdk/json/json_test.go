package opsjson

import (
	"testing"
)

func TestEncode_Struct(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	result, err := Encode(Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Size == 0 {
		t.Fatal("expected non-zero size")
	}
	if result.JSON == "" {
		t.Fatal("expected non-empty JSON")
	}
	// Verify it contains expected fields
	if !contains(result.JSON, "Alice") || !contains(result.JSON, "30") {
		t.Fatalf("JSON missing expected values: %s", result.JSON)
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
	if result.Size == 0 {
		t.Fatal("expected non-zero size")
	}
	if !contains(result.JSON, "key") || !contains(result.JSON, "value") {
		t.Fatalf("JSON missing expected values: %s", result.JSON)
	}
}

func TestEncode_Slice(t *testing.T) {
	data := []string{"a", "b", "c"}
	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(result.JSON, "a") || !contains(result.JSON, "c") {
		t.Fatalf("JSON missing expected values: %s", result.JSON)
	}
}

func TestEncode_String(t *testing.T) {
	result, err := Encode("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JSON != `"hello"` {
		t.Fatalf("expected quoted string, got: %s", result.JSON)
	}
}

func TestEncode_Number(t *testing.T) {
	result, err := Encode(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JSON != "42" {
		t.Fatalf("expected 42, got: %s", result.JSON)
	}
}

func TestEncode_Unsupported(t *testing.T) {
	_, err := Encode(make(chan int))
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestDecode_ValidObject(t *testing.T) {
	input := `{"name":"Alice","age":30}`
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
	// JSON numbers decode as float64
	if m["age"] != float64(30) {
		t.Fatalf("expected 30, got %v", m["age"])
	}
}

func TestDecode_Array(t *testing.T) {
	input := `[1, 2, 3]`
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

func TestDecode_String(t *testing.T) {
	result, err := Decode(`"hello"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data != "hello" {
		t.Fatalf("expected hello, got %v", result.Data)
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

func TestDecode_InvalidJSON(t *testing.T) {
	_, err := Decode("{invalid")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDecode_Empty(t *testing.T) {
	_, err := Decode("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestEncodeDecode_Roundtrip(t *testing.T) {
	original := map[string]interface{}{
		"key":   "value",
		"count": float64(42),
		"nested": map[string]interface{}{
			"flag": true,
		},
	}
	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, err := Decode(encoded.JSON)
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

// contains checks if substr is within s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
