package patch

import (
	"os"
	"strings"
	"testing"
)

func TestApply_EmptyPatch(t *testing.T) {
	_, err := Apply("", false)
	if err == nil {
		t.Error("expected error for empty patch")
	}
}

func TestDryRun_EmptyPatch(t *testing.T) {
	_, err := DryRun("")
	if err == nil {
		t.Error("expected error for empty patch")
	}
}

func TestParsePatch_NoTargetFile(t *testing.T) {
	_, _, err := parsePatch("no diff here\njust text\n")
	if err == nil {
		t.Error("expected error for patch without +++ line")
	}
}

func TestApplyAndReverse(t *testing.T) {
	// Create a test file
	content := "line1\nline2\nline3\n"
	tmpFile := "/tmp/patch_test_apply.ops"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	patch := `--- a/test.txt
+++ b/` + tmpFile + `
@@ -1,3 +1,3 @@
 line1
-line2
+line2_modified
 line3
`

	// Apply forward
	result, err := Apply(patch, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied {
		t.Error("expected applied=true")
	}
	if result.Hunks != 1 {
		t.Errorf("expected 1 hunk, got %d", result.Hunks)
	}

	// Verify content changed
	newContent, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(newContent), "line2_modified") {
		t.Error("expected line2_modified in output")
	}

	// Reverse
	result2, err := Apply(patch, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result2.Applied {
		t.Error("expected applied=true on reverse")
	}
	if !result2.Reversed {
		t.Error("expected reversed=true")
	}

	// Verify content restored
	restored, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(restored), "line2\n") || strings.Contains(string(restored), "line2_modified") {
		t.Errorf("expected original content restored, got: %s", string(restored))
	}
}

func TestDryRun_FileNotExists(t *testing.T) {
	patch := `--- a/test.txt
+++ b/nonexistent_file_xyz.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2_modified
 line3
`
	result, err := DryRun(patch)
	if err != nil {
		t.Fatal(err)
	}
	if result.CanApply {
		t.Error("expected can_apply=false for nonexistent file")
	}
}

func TestDryRun_CanApply(t *testing.T) {
	tmpFile := "/tmp/patch_dryrun_test.txt"
	os.WriteFile(tmpFile, []byte("line1\nline2\nline3\n"), 0644)
	defer os.Remove(tmpFile)

	patch := `--- a/test.txt
+++ ` + tmpFile + `
@@ -1,3 +1,3 @@
 line1
-line2
+line2_modified
 line3
`
	result, err := DryRun(patch)
	if err != nil {
		t.Fatal(err)
	}
	if !result.CanApply {
		t.Errorf("expected can_apply=true, got message: %s", result.Message)
	}
	if result.Hunks != 1 {
		t.Errorf("expected 1 hunk, got %d", result.Hunks)
	}
}

func TestResultFields(t *testing.T) {
	r := Result{
		Applied:  true,
		Reversed: false,
		File:     "/tmp/test.txt",
		Hunks:    2,
		Message:  "applied 2 hunks",
	}
	if !r.Applied {
		t.Error("applied should be true")
	}
	if r.Hunks != 2 {
		t.Error("hunks should be 2")
	}
}

func TestDryRunResultFields(t *testing.T) {
	r := DryRunResult{
		CanApply: true,
		File:     "/tmp/test.txt",
		Hunks:    1,
		Message:  "ok",
	}
	if !r.CanApply {
		t.Error("can_apply should be true")
	}
}
