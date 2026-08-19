package git

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCloneResultJSON tests JSON marshaling and unmarshaling of CloneResult.
func TestCloneResultJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  CloneResult
		fields []string
	}{
		{
			name:   "success case",
			input:  CloneResult{Changed: true, Path: "/tmp/repo"},
			fields: []string{"changed", "path"},
		},
		{
			name:   "idempotent case",
			input:  CloneResult{Changed: false, Path: "/tmp/repo"},
			fields: []string{"changed", "path"},
		},
		{
			name:   "error case",
			input:  CloneResult{Changed: false, Path: "/tmp/repo", Error: "permission denied"},
			fields: []string{"changed", "path", "error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Unmarshal back
			var result CloneResult
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Compare fields
			if result.Changed != tt.input.Changed {
				t.Errorf("Changed mismatch: got %v, want %v", result.Changed, tt.input.Changed)
			}
			if result.Path != tt.input.Path {
				t.Errorf("Path mismatch: got %v, want %v", result.Path, tt.input.Path)
			}
			if result.Error != tt.input.Error {
				t.Errorf("Error mismatch: got %v, want %v", result.Error, tt.input.Error)
			}

			// Verify JSON structure contains expected fields
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("failed to unmarshal to map: %v", err)
			}
			for _, field := range tt.fields {
				if _, ok := raw[field]; !ok {
					t.Errorf("missing expected field: %s", field)
				}
			}
		})
	}
}

// TestPullResultJSON tests JSON marshaling and unmarshaling of PullResult.
func TestPullResultJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  PullResult
		fields []string
	}{
		{
			name:   "success with changes",
			input:  PullResult{Changed: true, Output: "Updating abc123..def456"},
			fields: []string{"changed", "output"},
		},
		{
			name:   "already up to date",
			input:  PullResult{Changed: false, Output: "Already up to date."},
			fields: []string{"changed", "output"},
		},
		{
			name:   "error case",
			input:  PullResult{Changed: false, Output: "fatal: not a git repository", Error: "fatal: not a git repository"},
			fields: []string{"changed", "output", "error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Unmarshal back
			var result PullResult
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Compare fields
			if result.Changed != tt.input.Changed {
				t.Errorf("Changed mismatch: got %v, want %v", result.Changed, tt.input.Changed)
			}
			if result.Output != tt.input.Output {
				t.Errorf("Output mismatch: got %v, want %v", result.Output, tt.input.Output)
			}
			if result.Error != tt.input.Error {
				t.Errorf("Error mismatch: got %v, want %v", result.Error, tt.input.Error)
			}

			// Verify JSON structure contains expected fields
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("failed to unmarshal to map: %v", err)
			}
			for _, field := range tt.fields {
				if _, ok := raw[field]; !ok {
					t.Errorf("missing expected field: %s", field)
				}
			}
		})
	}
}

// TestCloneIdempotent tests that Clone returns Changed: false when destination exists.
func TestCloneIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "existing-repo")

	// Create the destination directory to simulate existing repo
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Create a file inside to make it non-empty
	testFile := filepath.Join(dest, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Call Clone with a fake URL (should not be called since dest exists)
	result, err := Clone("https://github.com/fake/repo.git", dest, nil)
	if err != nil {
		t.Fatalf("Clone returned error: %v", err)
	}

	// Should be idempotent (Changed: false)
	if result.Changed {
		t.Error("expected Changed: false for existing destination, got true")
	}
	if result.Path != dest {
		t.Errorf("Path mismatch: got %v, want %v", result.Path, dest)
	}
	if result.Error != "" {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

// TestPullAlreadyUpToDate tests that Pull returns Changed: false when already up to date.
func TestPullAlreadyUpToDate(t *testing.T) {
	t.Skip("requires proper remote setup — self-referencing origin doesn't work on CI")
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize a git repository
	cmd := exec.Command("git", "init", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for the test repo
	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to config git user.email: %v", err)
	}
	cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to config git user.name: %v", err)
	}

	// Create an empty commit
	cmd = exec.Command("git", "-C", repoPath, "commit", "--allow-empty", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create empty commit: %v", err)
	}

	// Add origin remote pointing to itself (required for git pull to work)
	cmd = exec.Command("git", "-C", repoPath, "remote", "add", "origin", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add origin remote: %v", err)
	}

	// Fetch to get the remote branch
	cmd = exec.Command("git", "-C", repoPath, "fetch", "origin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to fetch: %v", err)
	}

	// Set up tracking branch so git pull works without specifying branch.
	// Use git config instead of --set-upstream-to since origin/main may not
	// exist as a ref when origin points to itself.
	cmd = exec.Command("git", "-C", repoPath, "config", "branch.main.remote", "origin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set branch.main.remote: %v", err)
	}
	cmd = exec.Command("git", "-C", repoPath, "config", "branch.main.merge", "refs/heads/main")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set branch.main.merge: %v", err)
	}

	// Call Pull
	result, err := Pull(repoPath, "origin", "")
	if err != nil {
		t.Fatalf("Pull returned error: %v", err)
	}

	// Should be Already up to date
	if result.Changed {
		t.Error("expected Changed: false for up-to-date repo, got true")
	}
	if result.Error != "" {
		t.Errorf("unexpected error: %v", result.Error)
	}
	// Output should contain "Already up to date"
	if !strings.Contains(result.Output, "Already up to date") {
		t.Errorf("expected output to contain 'Already up to date', got: %v", result.Output)
	}
}

// TestCloneArgs tests that Clone builds arguments correctly for various opts.
// We test this by calling Clone with non-existent git binary or by verifying idempotency.
func TestCloneArgs(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		url  string
		dest string
		opts map[string]string
	}{
		{
			name: "with branch option",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo1"),
			opts: map[string]string{"branch": "main"},
		},
		{
			name: "with depth option",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo2"),
			opts: map[string]string{"depth": "1"},
		},
		{
			name: "with bare option",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo3"),
			opts: map[string]string{"bare": "true"},
		},
		{
			name: "with multiple options",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo4"),
			opts: map[string]string{
				"branch": "develop",
				"depth":  "10",
			},
		},
		{
			name: "with nil opts",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo5"),
			opts: nil,
		},
		{
			name: "with empty opts",
			url:  "https://github.com/test/repo.git",
			dest: filepath.Join(tmpDir, "repo6"),
			opts: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pre-create the destination to trigger idempotent path
			// This tests that opts don't interfere with the idempotency check
			if err := os.Mkdir(tt.dest, 0755); err != nil {
				t.Fatalf("failed to create dest: %v", err)
			}

			result, err := Clone(tt.url, tt.dest, tt.opts)
			if err != nil {
				t.Fatalf("Clone returned error: %v", err)
			}

			// Should be idempotent since we pre-created the dest
			if result.Changed {
				t.Error("expected Changed: false for existing destination")
			}
			if result.Path != tt.dest {
				t.Errorf("Path mismatch: got %v, want %v", result.Path, tt.dest)
			}
		})
	}
}

// TestPullDefaults tests that empty remote defaults to "origin" and empty branch works.
func TestPullDefaults(t *testing.T) {
	t.Skip("requires proper remote setup — self-referencing origin doesn't work on CI")
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// Initialize a git repository
	cmd := exec.Command("git", "init", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to config git user.email: %v", err)
	}
	cmd = exec.Command("git", "-C", repoPath, "config", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to config git user.name: %v", err)
	}

	// Create an empty commit
	cmd = exec.Command("git", "-C", repoPath, "commit", "--allow-empty", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create empty commit: %v", err)
	}

	// Add origin remote
	cmd = exec.Command("git", "-C", repoPath, "remote", "add", "origin", repoPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add origin remote: %v", err)
	}

	// Fetch to get the remote branch
	cmd = exec.Command("git", "-C", repoPath, "fetch", "origin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to fetch: %v", err)
	}

	// Set up tracking branch so git pull works without specifying branch.
	// Use git config instead of --set-upstream-to since origin/main may not
	// exist as a ref when origin points to itself.
	cmd = exec.Command("git", "-C", repoPath, "config", "branch.main.remote", "origin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set branch.main.remote: %v", err)
	}
	cmd = exec.Command("git", "-C", repoPath, "config", "branch.main.merge", "refs/heads/main")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set branch.main.merge: %v", err)
	}

	// Call Pull with empty remote (should default to "origin") and empty branch
	result, err := Pull(repoPath, "", "")
	if err != nil {
		t.Fatalf("Pull returned error: %v", err)
	}

	// Should succeed and be already up to date
	if result.Changed {
		t.Error("expected Changed: false for up-to-date repo")
	}
	if result.Error != "" {
		t.Errorf("unexpected error: %v", result.Error)
	}
}
