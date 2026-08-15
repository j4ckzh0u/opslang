// Package opstime provides time-related operations for OpsLang.
// All functions return structured results with JSON tags, enabling easy
// serialization and downstream processing. Uses pure Go stdlib time.
package opstime

import (
	"fmt"
	"time"
)

// TimeInfo is returned by Now, holding current time in various formats.
type TimeInfo struct {
	Unix     int64  `json:"unix"`
	UnixNano int64  `json:"unix_nano"`
	ISO8601  string `json:"iso8601"`
	UTC      string `json:"utc"`
	Timezone string `json:"timezone"`
}

// FormatResult is returned by Format, holding the formatted time string.
type FormatResult struct {
	Formatted string `json:"formatted"`
}

// ParseResult is returned by Parse, holding the parsed time as Unix and ISO8601.
type ParseResult struct {
	Unix    int64  `json:"unix"`
	ISO8601 string `json:"iso8601"`
}

// DurationResult is returned by Since, holding the duration from a past timestamp to now.
type DurationResult struct {
	Seconds       float64 `json:"seconds"`
	Minutes       float64 `json:"minutes"`
	Hours         float64 `json:"hours"`
	HumanReadable string  `json:"human_readable"`
}

const (
	defaultLayout = "2006-01-02 15:04:05"
)

// Now returns the current time in various formats.
func Now() TimeInfo {
	now := time.Now()
	utc := now.UTC()
	return TimeInfo{
		Unix:     now.Unix(),
		UnixNano: now.UnixNano(),
		ISO8601:  now.Format(time.RFC3339),
		UTC:      utc.Format("2006-01-02T15:04:05Z"),
		Timezone: now.Location().String(),
	}
}

// Format converts a Unix timestamp to a formatted time string using the given layout.
// If layout is empty, defaults to "2006-01-02 15:04:05".
func Format(unix int64, layout string) (FormatResult, error) {
	if layout == "" {
		layout = defaultLayout
	}
	t := time.Unix(unix, 0)
	return FormatResult{
		Formatted: t.Format(layout),
	}, nil
}

// Parse parses a time string using the given layout and returns the Unix timestamp and ISO8601.
// If layout is empty, defaults to "2006-01-02 15:04:05".
func Parse(layout string, value string) (ParseResult, error) {
	if layout == "" {
		layout = defaultLayout
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		return ParseResult{}, fmt.Errorf("opstime.Parse: %w", err)
	}
	return ParseResult{
		Unix:    t.Unix(),
		ISO8601: t.Format(time.RFC3339),
	}, nil
}

// Since calculates the duration from the given Unix timestamp to now.
// Returns seconds, minutes, hours, and a human-readable string in "Xh Ym Zs" format.
func Since(unix int64) DurationResult {
	now := time.Now()
	past := time.Unix(unix, 0)
	d := now.Sub(past)

	totalSeconds := d.Seconds()
	totalMinutes := d.Minutes()
	totalHours := d.Hours()

	// Build human-readable format
	hours := int(totalHours)
	minutes := int(totalMinutes) % 60
	seconds := int(totalSeconds) % 60

	humanReadable := fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)

	return DurationResult{
		Seconds:       totalSeconds,
		Minutes:       totalMinutes,
		Hours:         totalHours,
		HumanReadable: humanReadable,
	}
}
