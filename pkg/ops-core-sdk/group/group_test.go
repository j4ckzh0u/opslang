package group

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleGroupContent is a realistic /etc/group fixture.
const sampleGroupContent = `# /etc/group: group database
root:x:0:
daemon:x:1:
bin:x:2:
sys:x:3:
adm:x:4:syslog,admin
staff:x:50:alice,bob,charlie
guests:x:100:
empty-gid:x:999:
`

// writeTempGroupFile writes content to a temp file and returns its path.
// The caller should defer os.Remove on the returned path.
func writeTempGroupFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "group")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp group file: %v", err)
	}
	return path
}

func TestInfo(t *testing.T) {
	path := writeTempGroupFile(t, sampleGroupContent)
	old := groupFile
	groupFile = path
	defer func() { groupFile = old }()

	t.Run("existing group with members", func(t *testing.T) {
		info, err := Info("staff")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "staff" {
			t.Errorf("Name: got %q, want %q", info.Name, "staff")
		}
		if info.GID != 50 {
			t.Errorf("GID: got %d, want %d", info.GID, 50)
		}
		if len(info.Members) != 3 {
			t.Fatalf("Members len: got %d, want 3", len(info.Members))
		}
		want := []string{"alice", "bob", "charlie"}
		for i, m := range want {
			if info.Members[i] != m {
				t.Errorf("Members[%d]: got %q, want %q", i, info.Members[i], m)
			}
		}
	})

	t.Run("existing group with no members", func(t *testing.T) {
		info, err := Info("root")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.GID != 0 {
			t.Errorf("GID: got %d, want %d", info.GID, 0)
		}
		if len(info.Members) != 0 {
			t.Errorf("Members: got %v, want empty", info.Members)
		}
	})

	t.Run("missing group", func(t *testing.T) {
		_, err := Info("nonexistent")
		if err == nil {
			t.Fatal("expected error for missing group, got nil")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should contain 'not found': %v", err)
		}
	})
}

func TestList(t *testing.T) {
	path := writeTempGroupFile(t, sampleGroupContent)
	old := groupFile
	groupFile = path
	defer func() { groupFile = old }()

	groups, err := List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 8 non-comment lines in the fixture
	if len(groups) != 8 {
		t.Fatalf("len: got %d, want 8", len(groups))
	}

	// First entry
	if groups[0].Name != "root" || groups[0].GID != 0 {
		t.Errorf("first group: got %+v, want root/0", groups[0])
	}

	// staff entry (index 5)
	staff := groups[5]
	if staff.Name != "staff" || staff.GID != 50 {
		t.Errorf("staff: got %+v, want staff/50", staff)
	}
	if len(staff.Members) != 3 {
		t.Errorf("staff members: got %d, want 3", len(staff.Members))
	}
}

func TestListEmptyFile(t *testing.T) {
	path := writeTempGroupFile(t, "")
	old := groupFile
	groupFile = path
	defer func() { groupFile = old }()

	groups, err := List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestExists(t *testing.T) {
	path := writeTempGroupFile(t, sampleGroupContent)
	old := groupFile
	groupFile = path
	defer func() { groupFile = old }()

	t.Run("existing group", func(t *testing.T) {
		res, err := Exists("staff")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Exists {
			t.Error("Exists: got false, want true")
		}
	})

	t.Run("missing group", func(t *testing.T) {
		res, err := Exists("nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Exists {
			t.Error("Exists: got true, want false")
		}
	})
}

func TestParseMembers(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"   ", 0},
		{"alice", 1},
		{"alice,bob", 2},
		{"alice,bob,charlie", 3},
		{"alice,,bob", 2}, // empty segment skipped
	}
	for _, tc := range tests {
		got := parseMembers(tc.input)
		if len(got) != tc.want {
			t.Errorf("parseMembers(%q): got len %d, want %d", tc.input, len(got), tc.want)
		}
	}
}

// JSON round-trip tests below do not touch the filesystem.

func TestGroupInfoJSON(t *testing.T) {
	info := GroupInfo{
		GID:     50,
		Name:    "staff",
		Members: []string{"alice", "bob"},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	jsonStr := string(data)
	for _, want := range []string{`"gid"`, `"name"`, `"members"`} {
		if !strings.Contains(jsonStr, want) {
			t.Errorf("JSON missing field %s: %s", want, jsonStr)
		}
	}

	var decoded GroupInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.GID != info.GID {
		t.Errorf("GID: got %d, want %d", decoded.GID, info.GID)
	}
	if decoded.Name != info.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, info.Name)
	}
	if len(decoded.Members) != len(info.Members) {
		t.Errorf("Members len: got %d, want %d", len(decoded.Members), len(info.Members))
	}
}

func TestAddResultJSON(t *testing.T) {
	res := AddResult{Changed: true, GID: 200}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"changed"`) || !strings.Contains(jsonStr, `"gid"`) {
		t.Errorf("JSON missing expected fields: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"error"`) {
		t.Errorf("empty Error should be omitted: %s", jsonStr)
	}

	var decoded AddResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !decoded.Changed || decoded.GID != 200 {
		t.Errorf("decoded: %+v", decoded)
	}
}

func TestAddResultJSONWithError(t *testing.T) {
	res := AddResult{Changed: false, Error: "permission denied"}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"error"`) {
		t.Errorf("JSON should include error field: %s", jsonStr)
	}
}

func TestRemoveResultJSON(t *testing.T) {
	res := RemoveResult{Changed: true}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	jsonStr := string(data)
	if strings.Contains(jsonStr, `"error"`) {
		t.Errorf("empty Error should be omitted: %s", jsonStr)
	}
}

func TestExistsResultJSON(t *testing.T) {
	res := ExistsResult{Exists: true}
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"exists"`) {
		t.Errorf("JSON missing 'exists' field: %s", string(data))
	}

	var decoded ExistsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !decoded.Exists {
		t.Error("decoded.Exists: got false, want true")
	}
}

func TestInfoFileNotFound(t *testing.T) {
	old := groupFile
	groupFile = "/nonexistent/path/group"
	defer func() { groupFile = old }()

	_, err := Info("root")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestListFileNotFound(t *testing.T) {
	old := groupFile
	groupFile = "/nonexistent/path/group"
	defer func() { groupFile = old }()

	_, err := List()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestExistsFileNotFound(t *testing.T) {
	old := groupFile
	groupFile = "/nonexistent/path/group"
	defer func() { groupFile = old }()

	_, err := Exists("root")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
