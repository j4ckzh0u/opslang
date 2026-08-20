package type_debug

import "testing"

func TestDebugNil(t *testing.T) {
	r := Debug(nil)
	if r.Type != "null" {
		t.Errorf("expected null, got %s", r.Type)
	}
	if !r.IsNull {
		t.Error("expected is_null true")
	}
}

func TestDebugString(t *testing.T) {
	r := Debug("hello")
	if r.Type != "string" {
		t.Errorf("expected string, got %s", r.Type)
	}
	if !r.IsString {
		t.Error("expected is_string true")
	}
	if r.Length != 5 {
		t.Errorf("expected length 5, got %d", r.Length)
	}
}

func TestDebugNumber(t *testing.T) {
	r := Debug(42)
	if r.Type != "number" {
		t.Errorf("expected number, got %s", r.Type)
	}
	if !r.IsNumber {
		t.Error("expected is_number true")
	}
}

func TestDebugBool(t *testing.T) {
	r := Debug(true)
	if r.Type != "bool" {
		t.Errorf("expected bool, got %s", r.Type)
	}
	if !r.IsBool {
		t.Error("expected is_bool true")
	}
}

func TestDebugList(t *testing.T) {
	r := Debug([]interface{}{1, "a", true})
	if r.Type != "list" {
		t.Errorf("expected list, got %s", r.Type)
	}
	if !r.IsList {
		t.Error("expected is_list true")
	}
	if r.Length != 3 {
		t.Errorf("expected length 3, got %d", r.Length)
	}
}

func TestDebugDict(t *testing.T) {
	r := Debug(map[string]interface{}{"a": 1, "b": "two"})
	if r.Type != "dict" {
		t.Errorf("expected dict, got %s", r.Type)
	}
	if !r.IsDict {
		t.Error("expected is_dict true")
	}
	if r.Length != 2 {
		t.Errorf("expected length 2, got %d", r.Length)
	}
	if len(r.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(r.Keys))
	}
}
