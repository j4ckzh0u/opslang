package set_fact

import "testing"

func TestSetAndGet(t *testing.T) {
	Clear()
	Set(map[string]interface{}{"a": 1, "b": "two"})
	v, ok := Get("a")
	if !ok || v != 1 {
		t.Errorf("expected a=1, got %v", v)
	}
	v, ok = Get("b")
	if !ok || v != "two" {
		t.Errorf("expected b=two, got %v", v)
	}
}

func TestGetAll(t *testing.T) {
	Clear()
	Set(map[string]interface{}{"x": 1})
	Set(map[string]interface{}{"y": 2})
	all := GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2, got %d", len(all))
	}
}

func TestClear(t *testing.T) {
	Clear()
	Set(map[string]interface{}{"a": 1})
	r := Clear()
	if len(r.AnsibleFacts) != 0 {
		t.Error("expected empty")
	}
	if len(GetAll()) != 0 {
		t.Error("expected empty after clear")
	}
}
