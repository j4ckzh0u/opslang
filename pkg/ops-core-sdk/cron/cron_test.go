package cron

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCronEntryJSON verifies CronEntry serializes correctly to JSON.
func TestCronEntryJSON(t *testing.T) {
	entry := CronEntry{
		Minute:     "0",
		Hour:       "*/2",
		DayOfMonth: "*",
		Month:      "*",
		DayOfWeek:  "1-5",
		Command:    "/usr/bin/backup.sh",
		User:       "root",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal CronEntry: %v", err)
	}

	var decoded CronEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CronEntry: %v", err)
	}

	if decoded != entry {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, entry)
	}

	// Verify JSON tags
	jsonStr := string(data)
	expectedFields := []string{
		`"minute"`, `"hour"`, `"day_of_month"`, `"month"`,
		`"day_of_week"`, `"command"`, `"user"`,
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing field %s: %s", field, jsonStr)
		}
	}
}

// TestCronEntryJSONOmitUser verifies the user field is omitted when empty.
func TestCronEntryJSONOmitUser(t *testing.T) {
	entry := CronEntry{
		Minute:     "30",
		Hour:       "6",
		DayOfMonth: "1",
		Month:      "*",
		DayOfWeek:  "*",
		Command:    "/usr/bin/monthly.sh",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal CronEntry: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, `"user"`) {
		t.Errorf("JSON output should omit empty user field: %s", jsonStr)
	}
}

// TestListResultJSON verifies ListResult serializes correctly.
func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Entries: []CronEntry{
			{Minute: "0", Hour: "0", DayOfMonth: "*", Month: "*", DayOfWeek: "*", Command: "/usr/bin/cleanup"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal ListResult: %v", err)
	}

	var decoded ListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ListResult: %v", err)
	}

	if len(decoded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(decoded.Entries))
	}
	if decoded.Entries[0].Command != "/usr/bin/cleanup" {
		t.Errorf("Command: got %q, want %q", decoded.Entries[0].Command, "/usr/bin/cleanup")
	}
}

// TestAddResultJSON verifies AddResult serializes correctly.
func TestAddResultJSON(t *testing.T) {
	result := AddResult{Changed: true}
	data, _ := json.Marshal(result)
	var decoded AddResult
	json.Unmarshal(data, &decoded)

	if decoded != result {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, result)
	}
}

// TestRemoveResultJSON verifies RemoveResult serializes correctly.
func TestRemoveResultJSON(t *testing.T) {
	result := RemoveResult{Changed: false, Error: "not found"}
	data, _ := json.Marshal(result)
	var decoded RemoveResult
	json.Unmarshal(data, &decoded)

	if decoded != result {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, result)
	}
}

// TestParseLine verifies single-line cron parsing.
func TestParseLine(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		entry  CronEntry
		ok     bool
	}{
		{
			name:  "standard entry",
			input: "0 */2 * * * /usr/bin/backup.sh",
			entry: CronEntry{
				Minute:     "0",
				Hour:       "*/2",
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "*",
				Command:    "/usr/bin/backup.sh",
			},
			ok: true,
		},
		{
			name:  "entry with multiple command arguments",
			input: "30 6 * * 1-5 /usr/bin/cmd --flag arg1 arg2",
			entry: CronEntry{
				Minute:     "30",
				Hour:       "6",
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "1-5",
				Command:    "/usr/bin/cmd --flag arg1 arg2",
			},
			ok: true,
		},
		{
			name:  "comment line",
			input: "# this is a comment",
			ok:    false,
		},
		{
			name:  "blank line",
			input: "",
			ok:    false,
		},
		{
			name:  "too few fields",
			input: "0 * *",
			ok:    false,
		},
		{
			name:  "leading whitespace",
			input: "  0 0 * * * /usr/bin/midnight.sh",
			entry: CronEntry{
				Minute:     "0",
				Hour:       "0",
				DayOfMonth: "*",
				Month:      "*",
				DayOfWeek:  "*",
				Command:    "/usr/bin/midnight.sh",
			},
			ok: true,
		},
		{
			name:  "environment variable line (fewer than 6 fields after splitting)",
			input: "SHELL=/bin/bash",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := parseLine(tt.input)
			if ok != tt.ok {
				t.Errorf("parseLine(%q) ok = %v, want %v", tt.input, ok, tt.ok)
				return
			}
			if !ok {
				return
			}
			if entry.Minute != tt.entry.Minute {
				t.Errorf("Minute: got %q, want %q", entry.Minute, tt.entry.Minute)
			}
			if entry.Hour != tt.entry.Hour {
				t.Errorf("Hour: got %q, want %q", entry.Hour, tt.entry.Hour)
			}
			if entry.DayOfMonth != tt.entry.DayOfMonth {
				t.Errorf("DayOfMonth: got %q, want %q", entry.DayOfMonth, tt.entry.DayOfMonth)
			}
			if entry.Month != tt.entry.Month {
				t.Errorf("Month: got %q, want %q", entry.Month, tt.entry.Month)
			}
			if entry.DayOfWeek != tt.entry.DayOfWeek {
				t.Errorf("DayOfWeek: got %q, want %q", entry.DayOfWeek, tt.entry.DayOfWeek)
			}
			if entry.Command != tt.entry.Command {
				t.Errorf("Command: got %q, want %q", entry.Command, tt.entry.Command)
			}
		})
	}
}

// TestEntryToLine verifies converting an entry back to a cron line.
func TestEntryToLine(t *testing.T) {
	entry := CronEntry{
		Minute:     "15",
		Hour:       "3",
		DayOfMonth: "1",
		Month:      "*/3",
		DayOfWeek:  "*",
		Command:    "/usr/bin/quarterly.sh",
	}

	want := "15 3 1 */3 * /usr/bin/quarterly.sh"
	got := entryToLine(entry)

	if got != want {
		t.Errorf("entryToLine() = %q, want %q", got, want)
	}
}

// fakeCrontabState stores the state for a fake crontab binary.
type fakeCrontabState struct {
	// userCrontabs maps user -> current crontab content.
	userCrontabs map[string]string
}

// newFakeCrontab creates a fake crontab binary in tmpDir and returns
// a state struct plus the path to the binary.
func newFakeCrontab(t *testing.T, tmpDir string) *fakeCrontabState {
	t.Helper()
	state := &fakeCrontabState{
		userCrontabs: make(map[string]string),
	}

	// Write the initial state to a file the script can read/write.
	stateFile := filepath.Join(tmpDir, "crontab_state")
	if err := os.WriteFile(stateFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create state file: %v", err)
	}

	script := `#!/bin/sh
STATE_FILE="` + stateFile + `"

# Parse arguments.
ACTION=""
USER=""
READ_STDIN=false

while [ $# -gt 0 ]; do
  case "$1" in
    -l) ACTION="list" ;;
    -u) shift; USER="$1" ;;
    -)  READ_STDIN=true ;;
    *)  ;;
  esac
  shift
done

if [ "$ACTION" = "list" ]; then
  # Look up user's crontab from state file.
  if [ -f "$STATE_FILE" ]; then
    CONTENT=$(grep "^${USER}:" "$STATE_FILE" 2>/dev/null | head -1 | sed "s/^${USER}://")
    if [ -z "$CONTENT" ]; then
      echo "no crontab for $USER" >&2
      exit 1
    fi
    # Decode literal \n back to real newlines.
    printf '%b\n' "$CONTENT"
  else
    echo "no crontab for $USER" >&2
    exit 1
  fi
  exit 0
fi

if [ "$READ_STDIN" = true ]; then
  # Read new crontab from stdin.
  NEW_CONTENT=$(cat)
  # Remove old entry for user.
  if [ -f "$STATE_FILE" ]; then
    grep -v "^${USER}:" "$STATE_FILE" > "${STATE_FILE}.tmp" 2>/dev/null || true
    mv "${STATE_FILE}.tmp" "$STATE_FILE"
  fi
  # Append new entry (replace newlines with literal \n for storage).
  ENCODED=$(echo "$NEW_CONTENT" | awk '{printf "%s\\n", $0}')
  echo "${USER}:${ENCODED}" >> "$STATE_FILE"
  exit 0
fi

echo "Unknown operation" >&2
exit 1
`
	binPath := filepath.Join(tmpDir, "crontab")
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake crontab: %v", err)
	}

	return state
}

// setFakeCrontab initializes the state with a crontab for the given user.
// This is a helper used directly in tests by writing to the state file.
func setFakeCrontab(t *testing.T, tmpDir, user, content string) {
	t.Helper()
	stateFile := filepath.Join(tmpDir, "crontab_state")

	// Read existing state, remove old entry for this user, write back.
	existing, _ := os.ReadFile(stateFile)
	lines := strings.Split(string(existing), "\n")
	var kept []string
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, user+":") {
			continue
		}
		kept = append(kept, line)
	}

	// Encode new content: replace newlines so it fits on one line.
	encoded := strings.ReplaceAll(content, "\n", "\\n")
	kept = append(kept, user+":"+encoded)

	if err := os.WriteFile(stateFile, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatalf("failed to update state file: %v", err)
	}
}

// TestListWithFakeCrontab tests the full List flow using a fake crontab.
func TestListWithFakeCrontab(t *testing.T) {
	tmpDir := t.TempDir()
	state := newFakeCrontab(t, tmpDir)
	_ = state

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	// Set up a crontab for "testuser".
	crontabContent := "# my crontab\n0 */2 * * * /usr/bin/backup.sh\n30 6 * * 1-5 /usr/bin/report.sh --full\n"
	setFakeCrontab(t, tmpDir, "testuser", crontabContent)

	result, err := List("testuser")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}

	// Check first entry.
	e0 := result.Entries[0]
	if e0.Minute != "0" || e0.Hour != "*/2" {
		t.Errorf("first entry time fields: got %s %s, want 0 */2", e0.Minute, e0.Hour)
	}
	if e0.Command != "/usr/bin/backup.sh" {
		t.Errorf("first entry Command: got %q, want %q", e0.Command, "/usr/bin/backup.sh")
	}
	if e0.User != "testuser" {
		t.Errorf("first entry User: got %q, want %q", e0.User, "testuser")
	}

	// Check second entry.
	e1 := result.Entries[1]
	if e1.Command != "/usr/bin/report.sh --full" {
		t.Errorf("second entry Command: got %q, want %q", e1.Command, "/usr/bin/report.sh --full")
	}
	if e1.DayOfWeek != "1-5" {
		t.Errorf("second entry DayOfWeek: got %q, want %q", e1.DayOfWeek, "1-5")
	}
}

// TestListEmptyCrontab verifies List returns an empty slice for a user with no crontab.
func TestListEmptyCrontab(t *testing.T) {
	tmpDir := t.TempDir()
	newFakeCrontab(t, tmpDir)

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	result, err := List("emptyuser")
	if err != nil {
		t.Fatalf("List returned error for user with no crontab: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

// TestListMissingCrontab verifies List returns an error when crontab binary is missing.
func TestListMissingCrontab(t *testing.T) {
	original := crontabBin
	crontabBin = "/nonexistent/path/to/crontab"
	defer func() { crontabBin = original }()

	_, err := List("testuser")
	if err == nil {
		t.Fatal("expected error when crontab binary is missing, got nil")
	}
}

// TestAddWithFakeCrontab tests adding an entry.
func TestAddWithFakeCrontab(t *testing.T) {
	tmpDir := t.TempDir()
	newFakeCrontab(t, tmpDir)

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	entry := CronEntry{
		Minute:     "0",
		Hour:       "0",
		DayOfMonth: "*",
		Month:      "*",
		DayOfWeek:  "*",
		Command:    "/usr/bin/midnight.sh",
	}

	result, err := Add("newuser", entry)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if !result.Changed {
		t.Error("Add: expected Changed=true, got false")
	}
}

// TestAddToExistingCrontab tests adding to a user with existing entries.
func TestAddToExistingCrontab(t *testing.T) {
	tmpDir := t.TempDir()
	newFakeCrontab(t, tmpDir)

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	existing := "0 */2 * * * /usr/bin/backup.sh\n"
	setFakeCrontab(t, tmpDir, "existinguser", existing)

	entry := CronEntry{
		Minute:     "30",
		Hour:       "6",
		DayOfMonth: "*",
		Month:      "*",
		DayOfWeek:  "1-5",
		Command:    "/usr/bin/weekday-report.sh",
	}

	result, err := Add("existinguser", entry)
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if !result.Changed {
		t.Error("Add: expected Changed=true, got false")
	}

	// Verify the entry was added by listing.
	listResult, err := List("existinguser")
	if err != nil {
		t.Fatalf("List after Add returned error: %v", err)
	}
	if len(listResult.Entries) < 2 {
		t.Errorf("expected at least 2 entries after Add, got %d", len(listResult.Entries))
	}
}

// TestRemoveWithFakeCrontab tests removing entries.
func TestRemoveWithFakeCrontab(t *testing.T) {
	tmpDir := t.TempDir()
	newFakeCrontab(t, tmpDir)

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	existing := "# comment\n0 */2 * * * /usr/bin/backup.sh\n30 6 * * 1-5 /usr/bin/report.sh\n"
	setFakeCrontab(t, tmpDir, "removeuser", existing)

	result, err := Remove("removeuser", "backup.sh")
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if !result.Changed {
		t.Error("Remove: expected Changed=true, got false")
	}
}

// TestRemoveNoMatch tests that Remove returns Changed=false when no lines match.
func TestRemoveNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	newFakeCrontab(t, tmpDir)

	original := crontabBin
	crontabBin = filepath.Join(tmpDir, "crontab")
	defer func() { crontabBin = original }()

	existing := "0 */2 * * * /usr/bin/backup.sh\n"
	setFakeCrontab(t, tmpDir, "nomatchuser", existing)

	result, err := Remove("nomatchuser", "nonexistent-command")
	if err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if result.Changed {
		t.Error("Remove: expected Changed=false when no lines match, got true")
	}
}

// TestRemoveMissingCrontab verifies Remove returns an error when crontab binary is missing.
func TestRemoveMissingCrontab(t *testing.T) {
	original := crontabBin
	crontabBin = "/nonexistent/path/to/crontab"
	defer func() { crontabBin = original }()

	result, err := Remove("testuser", "some-command")
	if err == nil {
		t.Fatal("expected error when crontab binary is missing, got nil")
	}
	if result.Changed {
		t.Error("Remove: expected Changed=false on error")
	}
	if result.Error == "" {
		t.Error("Remove: expected non-empty Error on failure")
	}
}

// TestAddMissingCrontab verifies Add returns an error when crontab binary is missing.
func TestAddMissingCrontab(t *testing.T) {
	original := crontabBin
	crontabBin = "/nonexistent/path/to/crontab"
	defer func() { crontabBin = original }()

	result, err := Add("testuser", CronEntry{
		Minute: "0", Hour: "0", DayOfMonth: "*", Month: "*", DayOfWeek: "*",
		Command: "/usr/bin/test.sh",
	})
	if err == nil {
		t.Fatal("expected error when crontab binary is missing, got nil")
	}
	if result.Changed {
		t.Error("Add: expected Changed=false on error")
	}
	if result.Error == "" {
		t.Error("Add: expected non-empty Error on failure")
	}
}
