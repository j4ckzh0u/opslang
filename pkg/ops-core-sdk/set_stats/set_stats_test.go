package set_stats

import "testing"

func TestSetEmpty(t *testing.T) {
	r := Set(nil)
	if r.Error == "" {
		t.Error("expected error for nil data")
	}
	r = Set(map[string]string{})
	if r.Error == "" {
		t.Error("expected error for empty data")
	}
}

func TestSetAndGet(t *testing.T) {
	Clear()
	r := Set(map[string]string{"deployed": "true", "version": "1.2.3"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}

	v, ok := Get("deployed")
	if !ok || v != "true" {
		t.Errorf("unexpected: %v %v", v, ok)
	}

	all := GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 stats, got %d", len(all))
	}
}

func TestClear(t *testing.T) {
	Set(map[string]string{"key": "val"})
	Clear()
	if len(GetAll()) != 0 {
		t.Error("expected empty after clear")
	}
}
