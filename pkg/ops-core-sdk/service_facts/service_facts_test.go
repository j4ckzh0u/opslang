package service_facts

import (
	"runtime"
	"testing"
)

func TestCollectRequiresLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		r := Collect()
		if r.Error == "" {
			t.Error("expected error on non-linux")
		}
	}
}

func TestJSON(t *testing.T) {
	r := ServiceFactsResult{Services: map[string]ServiceInfo{"ssh": {Name: "ssh", Status: "running"}}}
	s := r.JSON()
	if s == "" {
		t.Error("expected non-empty JSON")
	}
}
