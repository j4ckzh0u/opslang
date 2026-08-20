package maven_artifact

import (
	"os"
	"testing"
)

// TestBuildArtifactURL tests URL construction.
func TestBuildArtifactURL(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		groupID  string
		artifactID string
		version  string
		ext      string
		want     string
	}{
		{
			name: "standard jar",
			repoURL: "https://repo1.maven.org/maven2",
			groupID: "org.apache.commons",
			artifactID: "commons-lang3",
			version: "3.12.0",
			ext: "",
			want: "https://repo1.maven.org/maven2/org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
		},
		{
			name: "war file",
			repoURL: "https://repo.example.com/releases",
			groupID: "com.example",
			artifactID: "webapp",
			version: "1.0",
			ext: "war",
			want: "https://repo.example.com/releases/com/example/webapp/1.0/webapp-1.0.war",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArtifactURL(tt.repoURL, tt.groupID, tt.artifactID, tt.version, tt.ext)
			if got != tt.want {
				t.Errorf("buildArtifactURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDownloadValidation tests input validation.
func TestDownloadValidation(t *testing.T) {
	_, err := Download("", "g", "a", "v", "/tmp/dst", "jar")
	if err == nil {
		t.Fatal("expected error for empty repo_url, got nil")
	}
}

// TestResolveValidation tests input validation.
func TestResolveValidation(t *testing.T) {
	_, err := Resolve("", "g", "a", "v", "jar")
	if err == nil {
		t.Fatal("expected error for empty repo_url, got nil")
	}
}

// TestDeployValidation tests input validation.
func TestDeployValidation(t *testing.T) {
	_, err := Deploy("", "g", "a", "v", "/tmp/src", "jar")
	if err == nil {
		t.Fatal("expected error for empty repo_url, got nil")
	}
}

// TestGetLatestVersionValidation tests input validation.
func TestGetLatestVersionValidation(t *testing.T) {
	_, err := GetLatestVersion("", "g", "a")
	if err == nil {
		t.Fatal("expected error for empty repo_url, got nil")
	}
}

// TestChecksumValidation tests input validation.
func TestChecksumValidation(t *testing.T) {
	_, err := Checksum("")
	if err == nil {
		t.Fatal("expected error for empty file_path, got nil")
	}
}

// TestChecksumComputation tests actual checksum computation.
func TestChecksumComputation(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	os.WriteFile(tmpFile, []byte("hello world"), 0644)

	sum, err := Checksum(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum == "" {
		t.Fatal("expected non-empty checksum")
	}
	// SHA256 of "hello world" is known
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if sum != expected {
		t.Errorf("checksum = %q, want %q", sum, expected)
	}
}

// TestChecksumNonexistent tests checksum of non-existent file.
func TestChecksumNonexistent(t *testing.T) {
	_, err := Checksum("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}
