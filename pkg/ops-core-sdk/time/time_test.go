package opstime

import (
	"strings"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	before := time.Now().Unix()
	result := Now()
	after := time.Now().Unix()

	if result.Unix < before || result.Unix > after {
		t.Fatalf("Unix timestamp out of range: %d (expected between %d and %d)", result.Unix, before, after)
	}
	if result.UnixNano == 0 {
		t.Fatal("expected non-zero UnixNano")
	}
	if result.ISO8601 == "" {
		t.Fatal("expected non-empty ISO8601")
	}
	if !strings.Contains(result.ISO8601, "T") {
		t.Fatalf("ISO8601 format unexpected: %s", result.ISO8601)
	}
	if result.UTC == "" {
		t.Fatal("expected non-empty UTC")
	}
	if !strings.HasSuffix(result.UTC, "Z") {
		t.Fatalf("UTC should end with Z: %s", result.UTC)
	}
	if result.Timezone == "" {
		t.Fatal("expected non-empty Timezone")
	}
}

func TestFormat_DefaultLayout(t *testing.T) {
	// 2023-06-15 12:30:45 UTC
	unix := int64(1686832245)
	result, err := Format(unix, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Formatted == "" {
		t.Fatal("expected non-empty formatted string")
	}
	// Should contain date-like components
	if !strings.Contains(result.Formatted, "-") {
		t.Fatalf("expected date format with dashes: %s", result.Formatted)
	}
}

func TestFormat_CustomLayout(t *testing.T) {
	unix := int64(1686832245)
	result, err := Format(unix, "2006/01/02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Formatted, "/") {
		t.Fatalf("expected slashes in format: %s", result.Formatted)
	}
}

func TestFormat_KnownTimestamp(t *testing.T) {
	// 2024-01-01 00:00:00 UTC = 1704067200
	unix := int64(1704067200)
	result, err := Format(unix, "2006-01-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The exact output depends on timezone, but should contain "2024" and either
	// "01-01" or "01-02" depending on the local timezone offset
	if !strings.Contains(result.Formatted, "2024") {
		t.Fatalf("expected year 2024: %s", result.Formatted)
	}
}

func TestFormat_RFC3339(t *testing.T) {
	unix := int64(1686832245)
	result, err := Format(unix, time.RFC3339)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Formatted, "T") {
		t.Fatalf("expected RFC3339 format with T: %s", result.Formatted)
	}
}

func TestParse_DefaultLayout(t *testing.T) {
	result, err := Parse("", "2023-06-15 12:30:45")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Unix == 0 {
		t.Fatal("expected non-zero Unix timestamp")
	}
	if result.ISO8601 == "" {
		t.Fatal("expected non-empty ISO8601")
	}
	if !strings.Contains(result.ISO8601, "T") {
		t.Fatalf("ISO8601 format unexpected: %s", result.ISO8601)
	}
}

func TestParse_CustomLayout(t *testing.T) {
	result, err := Parse("2006/01/02", "2023/06/15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Unix == 0 {
		t.Fatal("expected non-zero Unix timestamp")
	}
}

func TestParse_InvalidInput(t *testing.T) {
	_, err := Parse("2006-01-02", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date input")
	}
}

func TestParse_InvalidLayout(t *testing.T) {
	_, err := Parse("not-a-layout", "2023-06-15")
	if err == nil {
		t.Fatal("expected error for invalid layout")
	}
}

func TestSince_PastTimestamp(t *testing.T) {
	// 1 hour ago
	oneHourAgo := time.Now().Add(-1 * time.Hour).Unix()
	result := Since(oneHourAgo)

	if result.Seconds <= 0 {
		t.Fatal("expected positive seconds")
	}
	if result.Minutes <= 0 {
		t.Fatal("expected positive minutes")
	}
	if result.Hours <= 0 {
		t.Fatal("expected positive hours")
	}
	if result.HumanReadable == "" {
		t.Fatal("expected non-empty HumanReadable")
	}
	// Should contain "h" and "m" and "s"
	if !strings.Contains(result.HumanReadable, "h") {
		t.Fatalf("expected hours in human readable: %s", result.HumanReadable)
	}
	if !strings.Contains(result.HumanReadable, "m") {
		t.Fatalf("expected minutes in human readable: %s", result.HumanReadable)
	}
	if !strings.Contains(result.HumanReadable, "s") {
		t.Fatalf("expected seconds in human readable: %s", result.HumanReadable)
	}
	// Hours should be approximately 1
	if result.Hours < 0.9 || result.Hours > 1.1 {
		t.Fatalf("expected ~1 hour, got %f", result.Hours)
	}
}

func TestSince_OlderTimestamp(t *testing.T) {
	// 2 days ago
	twoDaysAgo := time.Now().Add(-48 * time.Hour).Unix()
	result := Since(twoDaysAgo)

	if result.Hours < 47 || result.Hours > 49 {
		t.Fatalf("expected ~48 hours, got %f", result.Hours)
	}
}

func TestSince_ZeroTimestamp(t *testing.T) {
	// Unix epoch (1970-01-01)
	result := Since(0)
	if result.Seconds <= 0 {
		t.Fatal("expected positive seconds for epoch timestamp")
	}
	if result.Hours <= 0 {
		t.Fatal("expected positive hours for epoch timestamp")
	}
}
