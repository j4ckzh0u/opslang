package package_facts

import (
	"runtime"
	"testing"
)

func TestCollectRequiresLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		r := Collect(nil)
		if r.Error == "" {
			t.Error("expected error on non-linux")
		}
	}
}

func TestJSON(t *testing.T) {
	r := PackageFactsResult{Packages: map[string][]PackageInfo{"foo": {{Name: "foo", Version: "1.0"}}}}
	s := r.JSON()
	if s == "" {
		t.Error("expected non-empty JSON")
	}
}
