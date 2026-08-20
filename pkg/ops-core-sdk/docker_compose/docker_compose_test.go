package docker_compose

import (
	"testing"
)

func TestFindCompose(t *testing.T) {
	bin, err := findCompose()
	if err != nil {
		t.Skipf("docker compose not available: %v", err)
	}
	if bin == "" {
		t.Error("expected non-empty binary path")
	}
}

func TestStatusNoCompose(t *testing.T) {
	// Status will fail gracefully when no compose project exists
	_, _ = Status("")
}
