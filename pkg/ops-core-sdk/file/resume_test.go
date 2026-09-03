package file

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestPartialPaths(t *testing.T) {
	partPath, metadataPath, err := partialPaths("/tmp/archive.tar")
	if err != nil {
		t.Fatalf("partialPaths: %v", err)
	}
	if partPath != "/tmp/archive.tar.opslang.part" {
		t.Fatalf("part path = %q", partPath)
	}
	if metadataPath != "/tmp/archive.tar.opslang.part.json" {
		t.Fatalf("metadata path = %q", metadataPath)
	}
	if _, _, err := partialPaths(""); err == nil {
		t.Fatal("empty final path must fail")
	}
}

func TestNewPartialMetadataUsesUniqueSessions(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.FixedZone("test", 8*60*60))
	first, err := newPartialMetadata(0, testSHA256, now)
	if err != nil {
		t.Fatalf("newPartialMetadata first: %v", err)
	}
	second, err := newPartialMetadata(0, strings.ToUpper(testSHA256), now)
	if err != nil {
		t.Fatalf("newPartialMetadata second: %v", err)
	}
	if first.SessionID == second.SessionID {
		t.Fatal("separate transfers must have separate session IDs")
	}
	if len(first.SessionID) != 32 {
		t.Fatalf("session ID length = %d, want 32", len(first.SessionID))
	}
	if first.SourceSHA256 != testSHA256 {
		t.Fatalf("checksum = %q, want normalized lowercase", first.SourceSHA256)
	}
	if first.UpdatedAt.Location() != time.UTC {
		t.Fatalf("updated location = %v, want UTC", first.UpdatedAt.Location())
	}
}

func TestPartialMetadataValidateForSource(t *testing.T) {
	base := PartialMetadata{
		Version:       partialMetadataVersion,
		SessionID:     "session",
		SourceSize:    10,
		SourceSHA256:  testSHA256,
		ConfirmedSize: 5,
		UpdatedAt:     time.Now().UTC(),
	}
	tests := []struct {
		name       string
		metadata   PartialMetadata
		sourceSize int64
		checksum   string
		partSize   int64
		wantErr    bool
	}{
		{name: "valid", metadata: base, sourceSize: 10, checksum: testSHA256, partSize: 5},
		{name: "unconfirmed tail", metadata: base, sourceSize: 10, checksum: testSHA256, partSize: 8},
		{name: "changed size", metadata: base, sourceSize: 11, checksum: testSHA256, partSize: 5, wantErr: true},
		{name: "changed checksum", metadata: base, sourceSize: 10, checksum: strings.Repeat("b", 64), partSize: 5, wantErr: true},
		{name: "short partial", metadata: base, sourceSize: 10, checksum: testSHA256, partSize: 4, wantErr: true},
		{name: "oversized partial", metadata: base, sourceSize: 10, checksum: testSHA256, partSize: 11, wantErr: true},
		{name: "negative partial", metadata: base, sourceSize: 10, checksum: testSHA256, partSize: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.validateForSource(tt.sourceSize, tt.checksum, tt.partSize)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateForSource error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPartialMetadataFieldValidation(t *testing.T) {
	valid := PartialMetadata{
		Version:      partialMetadataVersion,
		SessionID:    "session",
		SourceSHA256: testSHA256,
		UpdatedAt:    time.Now().UTC(),
	}
	tests := []struct {
		name   string
		mutate func(*PartialMetadata)
	}{
		{name: "version", mutate: func(m *PartialMetadata) { m.Version = 2 }},
		{name: "session", mutate: func(m *PartialMetadata) { m.SessionID = "" }},
		{name: "source size", mutate: func(m *PartialMetadata) { m.SourceSize = -1 }},
		{name: "checksum", mutate: func(m *PartialMetadata) { m.SourceSHA256 = "bad" }},
		{name: "confirmed size", mutate: func(m *PartialMetadata) { m.ConfirmedSize = 1 }},
		{name: "updated time", mutate: func(m *PartialMetadata) { m.UpdatedAt = time.Time{} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := valid
			tt.mutate(&metadata)
			if err := metadata.validateFields(); err == nil {
				t.Fatal("invalid metadata must fail validation")
			}
		})
	}
}

func TestPartialMetadataExpiredBoundary(t *testing.T) {
	updated := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	metadata := PartialMetadata{UpdatedAt: updated}
	retention := 24 * time.Hour
	if metadata.expired(updated.Add(retention-time.Nanosecond), retention) {
		t.Fatal("metadata expired before retention boundary")
	}
	if !metadata.expired(updated.Add(retention), retention) {
		t.Fatal("metadata must expire at retention boundary")
	}
	if metadata.expired(updated.Add(-time.Hour), retention) {
		t.Fatal("clock moving backwards must not expire metadata")
	}
	if metadata.expired(updated.Add(48*time.Hour), 0) {
		t.Fatal("non-positive retention disables expiry")
	}
}

func TestValidatePartialPrefix(t *testing.T) {
	source := bytes.NewReader([]byte("abcdefghij"))
	tests := []struct {
		name      string
		partial   *bytes.Reader
		confirmed int64
		wantErr   bool
	}{
		{name: "empty", partial: bytes.NewReader(nil), confirmed: 0},
		{name: "matching", partial: bytes.NewReader([]byte("abcdefghZZ")), confirmed: 8},
		{name: "mismatch", partial: bytes.NewReader([]byte("abcdXfgh")), confirmed: 8, wantErr: true},
		{name: "truncated", partial: bytes.NewReader([]byte("abc")), confirmed: 8, wantErr: true},
		{name: "oversized offset", partial: bytes.NewReader([]byte("abcdefghij")), confirmed: 11, wantErr: true},
		{name: "negative offset", partial: bytes.NewReader(nil), confirmed: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePartialPrefix(source, tt.partial, tt.confirmed)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePartialPrefix error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := validatePartialPrefix(nil, bytes.NewReader(nil), 0); err == nil {
		t.Fatal("nil source must fail")
	}
}

func TestWriteAndLoadPartialMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.part.json")
	metadata, err := newPartialMetadata(4, testSHA256, time.Now())
	if err != nil {
		t.Fatalf("newPartialMetadata: %v", err)
	}
	metadata.ConfirmedSize = 2
	if err := writePartialMetadata(path, metadata); err != nil {
		t.Fatalf("writePartialMetadata: %v", err)
	}
	loaded, err := loadPartialMetadata(path)
	if err != nil {
		t.Fatalf("loadPartialMetadata: %v", err)
	}
	if loaded != metadata {
		t.Fatalf("loaded metadata = %+v, want %+v", loaded, metadata)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat metadata: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("metadata mode = %o, want 600", info.Mode().Perm())
	}
}

func TestLoadPartialMetadataRejectsCorruptionAndEmptyPath(t *testing.T) {
	if _, err := loadPartialMetadata(""); err == nil {
		t.Fatal("empty metadata path must fail")
	}
	path := filepath.Join(t.TempDir(), "corrupt.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatalf("write corrupt metadata: %v", err)
	}
	if _, err := loadPartialMetadata(path); err == nil {
		t.Fatal("corrupt metadata must fail")
	}
}

func TestCommitPartialFile(t *testing.T) {
	dir := t.TempDir()
	partPath := filepath.Join(dir, "payload.part")
	finalPath := filepath.Join(dir, "payload")
	content := []byte("complete payload")
	if err := os.WriteFile(partPath, content, 0600); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	checksum, err := computeFileChecksum(partPath)
	if err != nil {
		t.Fatalf("checksum: %v", err)
	}
	if err := commitPartialFile(partPath, finalPath, int64(len(content)), checksum); err != nil {
		t.Fatalf("commitPartialFile: %v", err)
	}
	got, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("read final: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("final content = %q", got)
	}
	if _, err := os.Stat(partPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial path still exists or stat failed: %v", err)
	}
}

func TestCommitPartialFilePreservesFinalOnMismatch(t *testing.T) {
	dir := t.TempDir()
	partPath := filepath.Join(dir, "payload.part")
	finalPath := filepath.Join(dir, "payload")
	if err := os.WriteFile(partPath, []byte("partial"), 0600); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	if err := os.WriteFile(finalPath, []byte("stable"), 0600); err != nil {
		t.Fatalf("write final: %v", err)
	}
	if err := commitPartialFile(partPath, finalPath, 7, testSHA256); err == nil {
		t.Fatal("checksum mismatch must fail")
	}
	got, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("read final: %v", err)
	}
	if string(got) != "stable" {
		t.Fatalf("final content changed to %q", got)
	}
}
