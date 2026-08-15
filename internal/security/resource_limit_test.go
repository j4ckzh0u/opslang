package security

import (
	"runtime"
	"testing"
)

func TestApplyResourceLimit(t *testing.T) {
	limit := ResourceLimit{
		CPUPercent: 50,
		MemoryMB:   512,
	}

	applied, err := Apply(limit)
	if applied == nil {
		t.Fatal("Apply returned nil AppliedLimit")
	}

	// On non-Linux, we expect an error but still get a no-op AppliedLimit
	if runtime.GOOS != "linux" {
		if err == nil {
			t.Error("Expected error on non-Linux platform")
		}
		if applied.Method != "none" {
			t.Errorf("Method = %v, want 'none' on non-Linux", applied.Method)
		}
	}

	// Cleanup should not panic
	applied.Cleanup()

	// Cleanup should be idempotent
	applied.Cleanup()
}

func TestDefaultResourceLimit(t *testing.T) {
	limit := DefaultResourceLimit()

	if limit.CPUPercent <= 0 {
		t.Errorf("CPUPercent = %d, want > 0", limit.CPUPercent)
	}
	if limit.MemoryMB <= 0 {
		t.Errorf("MemoryMB = %d, want > 0", limit.MemoryMB)
	}
}

func TestResourceLimitString(t *testing.T) {
	limit := ResourceLimit{
		CPUPercent: 80,
		MemoryMB:   1024,
	}

	got := limit.String()
	expected := "CPU: 80%, Memory: 1024MB"
	if got != expected {
		t.Errorf("String() = %v, want %v", got, expected)
	}
}

func TestApplyZeroLimit(t *testing.T) {
	limit := ResourceLimit{
		CPUPercent: 0,
		MemoryMB:   0,
	}

	applied, err := Apply(limit)
	if applied == nil {
		t.Fatal("Apply returned nil AppliedLimit")
	}

	// Should still get a result (even if error on non-Linux)
	_ = err

	applied.Cleanup()
}

func TestApplyMultipleTimes(t *testing.T) {
	limit := ResourceLimit{
		CPUPercent: 50,
		MemoryMB:   256,
	}

	// Apply multiple times should not cause issues
	for i := 0; i < 3; i++ {
		applied, err := Apply(limit)
		if applied == nil {
			t.Fatalf("Apply returned nil on iteration %d", i)
		}
		_ = err
		applied.Cleanup()
	}
}
