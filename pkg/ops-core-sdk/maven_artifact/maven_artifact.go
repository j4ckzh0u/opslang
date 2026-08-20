// Package maven_artifact provides Maven artifact management operations.
// Equivalent to Ansible's maven_artifact module.
package maven_artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Result represents a generic artifact operation result.
type Result struct {
	Status     string `json:"status"`
	Changed    bool   `json:"changed"`
	Artifact   string `json:"artifact,omitempty"`
	Dest       string `json:"dest,omitempty"`
	Checksum   string `json:"checksum,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ArtifactInfo represents metadata about a Maven artifact.
type ArtifactInfo struct {
	Status   string `json:"status"`
	GroupID  string `json:"group_id"`
	ArtifactID string `json:"artifact_id"`
	Version  string `json:"version"`
	Extension string `json:"extension,omitempty"`
	URL      string `json:"url,omitempty"`
	Exists   bool   `json:"exists"`
}

// buildArtifactURL constructs the Maven repository URL for an artifact.
func buildArtifactURL(repoURL, groupID, artifactID, version, extension string) string {
	if extension == "" {
		extension = "jar"
	}
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	filename := fmt.Sprintf("%s-%s.%s", artifactID, version, extension)
	return fmt.Sprintf("%s/%s/%s/%s/%s", strings.TrimRight(repoURL, "/"), groupPath, artifactID, version, filename)
}

// Download downloads a Maven artifact to a local destination.
func Download(repoURL string, groupID string, artifactID string, version string, dest string, extension string) (Result, error) {
	if repoURL == "" || groupID == "" || artifactID == "" || version == "" || dest == "" {
		return Result{Status: "failed", Error: "repo_url, group_id, artifact_id, version, and dest are required"}, fmt.Errorf("required parameters missing")
	}

	url := buildArtifactURL(repoURL, groupID, artifactID, version, extension)
	artifactName := fmt.Sprintf("%s:%s:%s:%s", groupID, artifactID, version, extension)

	resp, err := http.Get(url)
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("download: %v", err)}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, url)}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Ensure dest directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("mkdir: %v", err)}, err
	}

	outFile, err := os.Create(dest)
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("create dest: %v", err)}, err
	}
	defer outFile.Close()

	hasher := sha256.New()
	written, err := io.Copy(outFile, io.TeeReader(resp.Body, hasher))
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("write: %v", err)}, err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return Result{
		Status:   "success",
		Changed:  true,
		Artifact: artifactName,
		Dest:     dest,
		Checksum: checksum,
		Size:     written,
	}, nil
}

// Resolve checks if an artifact exists in the repository.
func Resolve(repoURL string, groupID string, artifactID string, version string, extension string) (ArtifactInfo, error) {
	if repoURL == "" || groupID == "" || artifactID == "" || version == "" {
		return ArtifactInfo{Status: "failed"}, fmt.Errorf("repo_url, group_id, artifact_id, and version are required")
	}

	url := buildArtifactURL(repoURL, groupID, artifactID, version, extension)
	resp, err := http.Head(url)
	if err != nil {
		return ArtifactInfo{
			Status:     "failed",
			GroupID:    groupID,
			ArtifactID: artifactID,
			Version:    version,
			Exists:     false,
		}, nil
	}
	defer resp.Body.Close()

	exists := resp.StatusCode == http.StatusOK
	return ArtifactInfo{
		Status:     "success",
		GroupID:    groupID,
		ArtifactID: artifactID,
		Version:    version,
		Extension:  extension,
		URL:        url,
		Exists:     exists,
	}, nil
}

// Deploy uploads a local file to a Maven repository.
func Deploy(repoURL string, groupID string, artifactID string, version string, srcPath string, extension string) (Result, error) {
	if repoURL == "" || groupID == "" || artifactID == "" || version == "" || srcPath == "" {
		return Result{Status: "failed", Error: "repo_url, group_id, artifact_id, version, and src_path are required"}, fmt.Errorf("required parameters missing")
	}

	url := buildArtifactURL(repoURL, groupID, artifactID, version, extension)
	artifactName := fmt.Sprintf("%s:%s:%s:%s", groupID, artifactID, version, extension)

	file, err := os.Open(srcPath)
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("open src: %v", err)}, err
	}
	defer file.Close()

	stat, _ := file.Stat()
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("create request: %v", err)}, err
	}
	req.ContentLength = stat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("upload: %v", err)}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Result{Status: "failed", Artifact: artifactName, Error: fmt.Sprintf("HTTP %d", resp.StatusCode)}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return Result{
		Status:   "success",
		Changed:  true,
		Artifact: artifactName,
		Dest:     url,
		Size:     stat.Size(),
	}, nil
}

// GetLatestVersion queries a Maven repository for the latest version of an artifact.
// It checks the maven-metadata.xml file.
func GetLatestVersion(repoURL string, groupID string, artifactID string) (string, error) {
	if repoURL == "" || groupID == "" || artifactID == "" {
		return "", fmt.Errorf("repo_url, group_id, and artifact_id are required")
	}

	groupPath := strings.ReplaceAll(groupID, ".", "/")
	metaURL := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", strings.TrimRight(repoURL, "/"), groupPath, artifactID)

	resp, err := http.Get(metaURL)
	if err != nil {
		return "", fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching metadata", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read metadata: %w", err)
	}

	// Simple XML parsing — find <latest> or last <version> tag
	content := string(body)
	if idx := strings.Index(content, "<latest>"); idx >= 0 {
		end := strings.Index(content[idx:], "</latest>")
		if end > 0 {
			return strings.TrimSpace(content[idx+8 : idx+end]), nil
		}
	}

	// Fallback: find last <version> tag
	versions := []string{}
	for {
		idx := strings.Index(content, "<version>")
		if idx < 0 {
			break
		}
		end := strings.Index(content[idx:], "</version>")
		if end < 0 {
			break
		}
		v := strings.TrimSpace(content[idx+9 : idx+end])
		versions = append(versions, v)
		content = content[idx+end+10:]
	}

	if len(versions) > 0 {
		return versions[len(versions)-1], nil
	}

	return "", fmt.Errorf("no version found in metadata")
}

// Checksum computes the SHA256 checksum of a local file.
func Checksum(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
