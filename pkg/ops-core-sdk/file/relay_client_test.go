package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRelayFetchResumesAndCommitsAtomically(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.bin")
	content := []byte("0123456789abcdef")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	server, err := StartRelayHTTPServer(source, time.Minute, 2)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)

	destination := filepath.Join(t.TempDir(), "nested", "destination.bin")
	partPath, metadataPath, err := partialPaths(destination)
	if err != nil {
		t.Fatalf("partial paths: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatalf("create destination directory: %v", err)
	}
	const confirmed = 6
	if err := os.WriteFile(partPath, content[:confirmed], 0o600); err != nil {
		t.Fatalf("write partial file: %v", err)
	}
	metadata, err := newPartialMetadata(int64(len(content)), server.Info.SHA256, time.Now().UTC())
	if err != nil {
		t.Fatalf("new metadata: %v", err)
	}
	metadata.ConfirmedSize = confirmed
	if err := writePartialMetadata(metadataPath, metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	outcome, err := RelayFetch(context.Background(), relayFetchOptions(server, destination))
	if err != nil {
		t.Fatalf("relay fetch: %v", err)
	}
	if outcome.Status != "success" || outcome.ResumedBytes != confirmed || outcome.TransferredBytes != int64(len(content)-confirmed) {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("destination = %q, want %q", got, content)
	}
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		t.Fatalf("metadata remains after commit: %v", err)
	}

	skipped, err := RelayFetch(context.Background(), relayFetchOptions(server, destination))
	if err != nil {
		t.Fatalf("repeat relay fetch: %v", err)
	}
	if skipped.Status != "skipped" || skipped.TransferredBytes != 0 {
		t.Fatalf("repeat outcome: %+v", skipped)
	}
}

func TestRelayFetchRejectsFingerprintAndPreservesFinalFile(t *testing.T) {
	source := createRelayTestFile(t)
	server, err := StartRelayHTTPServer(source, time.Minute, 1)
	if err != nil {
		t.Fatalf("start relay server: %v", err)
	}
	defer stopRelayTestServer(t, server)
	destination := filepath.Join(t.TempDir(), "destination.bin")
	if err := os.WriteFile(destination, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write final file: %v", err)
	}
	opts := relayFetchOptions(server, destination)
	opts.CertFingerprint = strings.Repeat("0", 64)
	if _, err := RelayFetch(context.Background(), opts); err == nil || !strings.Contains(err.Error(), "fingerprint mismatch") {
		t.Fatalf("fingerprint error = %v", err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read final file: %v", err)
	}
	if string(got) != "keep" {
		t.Fatalf("final file changed to %q", got)
	}
}

func TestRelayFetchDecompressesAndVerifiesOriginal(t *testing.T) {
	original := createRelayTestFile(t)
	originalInfo, err := os.Stat(original)
	if err != nil {
		t.Fatal(err)
	}
	originalChecksum, err := computeFileChecksum(original)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := gzipFile(original)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(compressed)
	server, err := StartRelayHTTPServer(compressed, time.Minute, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRelayTestServer(t, server)
	destination := filepath.Join(t.TempDir(), "destination.bin")
	opts := relayFetchOptions(server, destination)
	opts.SHA256, opts.Size = originalChecksum, originalInfo.Size()
	opts.WireSHA256, opts.WireSize, opts.Decompress = server.Info.SHA256, server.Info.Size, true
	outcome, err := RelayFetch(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Status != "success" || outcome.Checksum != originalChecksum || outcome.Size != originalInfo.Size() {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("decompressed content mismatch")
	}
}

func TestRelayFetchResumesCompressedObject(t *testing.T) {
	original := createRelayTestFile(t)
	originalInfo, err := os.Stat(original)
	if err != nil {
		t.Fatal(err)
	}
	originalChecksum, err := computeFileChecksum(original)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := gzipFile(original)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(compressed)
	compressedBytes, err := os.ReadFile(compressed)
	if err != nil {
		t.Fatal(err)
	}
	server, err := StartRelayHTTPServer(compressed, time.Minute, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer stopRelayTestServer(t, server)

	destination := filepath.Join(t.TempDir(), "destination.bin")
	partPath, metadataPath, err := partialPaths(destination)
	if err != nil {
		t.Fatal(err)
	}
	confirmed := len(compressedBytes) / 2
	if err := os.WriteFile(partPath, compressedBytes[:confirmed], 0600); err != nil {
		t.Fatal(err)
	}
	metadata, err := newPartialMetadata(int64(len(compressedBytes)), server.Info.SHA256, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	metadata.ConfirmedSize = int64(confirmed)
	if err := writePartialMetadata(metadataPath, metadata); err != nil {
		t.Fatal(err)
	}

	opts := relayFetchOptions(server, destination)
	opts.SHA256, opts.Size = originalChecksum, originalInfo.Size()
	opts.WireSHA256, opts.WireSize, opts.Decompress = server.Info.SHA256, server.Info.Size, true
	outcome, err := RelayFetch(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Status != "success" || outcome.ResumedBytes != int64(confirmed) {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("compressed resume produced different content")
	}
}

func TestRelayFetchValidatesEmptyInput(t *testing.T) {
	if _, err := RelayFetch(context.Background(), RelayFetchOptions{}); err == nil {
		t.Fatal("expected empty input error")
	}
	if _, err := RelayFetch(nil, RelayFetchOptions{}); err == nil {
		t.Fatal("expected nil context error")
	}
}

func relayFetchOptions(server *RelayHTTPServer, destination string) RelayFetchOptions {
	return RelayFetchOptions{
		URL:             server.Info.URL,
		Token:           server.Info.Token,
		CertFingerprint: server.Info.CertFingerprint,
		SHA256:          server.Info.SHA256,
		Size:            server.Info.Size,
		Dest:            destination,
		PartRetention:   time.Hour,
		Timeout:         time.Second,
	}
}
