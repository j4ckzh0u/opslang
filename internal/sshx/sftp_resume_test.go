package sshx

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/sftp"
)

func writeRemoteFile(t *testing.T, client *sftp.Client, path string, content []byte) {
	t.Helper()
	file, err := client.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		t.Fatalf("open remote file %s: %v", path, err)
	}
	if _, err := file.Write(content); err != nil {
		closeErr := file.Close()
		t.Fatalf("write remote file %s: %v; close error: %v", path, err, closeErr)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close remote file %s: %v", path, err)
	}
}

func readRemoteFile(t *testing.T, client *sftp.Client, path string) []byte {
	t.Helper()
	file, err := client.Open(path)
	if err != nil {
		t.Fatalf("open remote file %s: %v", path, err)
	}
	content, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		t.Fatalf("read remote file %s: %v", path, readErr)
	}
	if closeErr != nil {
		t.Fatalf("close remote file %s: %v", path, closeErr)
	}
	return content
}

func TestSFTPClientStat(t *testing.T) {
	client, cleanup := sftpPipe(t)
	defer cleanup()
	writeRemoteFile(t, client, "stat.txt", []byte("data"))

	info, err := (&SFTPClient{client: client}).Stat(context.Background(), "stat.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() != 4 {
		t.Fatalf("size = %d, want 4", info.Size())
	}
}

func TestSFTPClientUploadAt(t *testing.T) {
	tests := []struct {
		name          string
		localContent  string
		remoteContent string
		offset        int64
		wantContent   string
		wantWritten   int64
		wantErr       bool
	}{
		{name: "zero offset replaces destination", localContent: "abcdefghij", remoteContent: "old trailing data", wantContent: "abcdefghij", wantWritten: 10},
		{name: "nonzero offset resumes and truncates tail", localContent: "abcdefghij", remoteContent: "abcXXXXXextra", offset: 3, wantContent: "abcdefghij", wantWritten: 7},
		{name: "empty source", localContent: "", remoteContent: "old", wantContent: "", wantWritten: 0},
		{name: "offset exceeds source", localContent: "abc", remoteContent: "abcdef", offset: 4, wantContent: "abcdef", wantErr: true},
		{name: "offset exceeds destination", localContent: "abcdef", remoteContent: "abc", offset: 4, wantContent: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := sftpPipe(t)
			defer cleanup()
			writeRemoteFile(t, client, "upload.part", []byte(tt.remoteContent))
			localPath := filepath.Join(t.TempDir(), "source")
			if err := os.WriteFile(localPath, []byte(tt.localContent), 0600); err != nil {
				t.Fatalf("write local source: %v", err)
			}

			written, err := (&SFTPClient{client: client}).UploadAt(context.Background(), localPath, "upload.part", tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UploadAt error = %v, wantErr %v", err, tt.wantErr)
			}
			if written != tt.wantWritten {
				t.Fatalf("written = %d, want %d", written, tt.wantWritten)
			}
			if got := string(readRemoteFile(t, client, "upload.part")); got != tt.wantContent {
				t.Fatalf("remote content = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestSFTPClientDownloadAt(t *testing.T) {
	tests := []struct {
		name        string
		remote      string
		local       string
		offset      int64
		wantContent string
		wantWritten int64
		wantErr     bool
	}{
		{name: "zero offset replaces destination", remote: "abcdefghij", local: "old trailing data", wantContent: "abcdefghij", wantWritten: 10},
		{name: "nonzero offset resumes and truncates tail", remote: "abcdefghij", local: "abcXXXXXextra", offset: 3, wantContent: "abcdefghij", wantWritten: 7},
		{name: "empty source", remote: "", local: "old", wantContent: "", wantWritten: 0},
		{name: "offset exceeds source", remote: "abc", local: "abcdef", offset: 4, wantContent: "abcdef", wantErr: true},
		{name: "offset exceeds destination", remote: "abcdef", local: "abc", offset: 4, wantContent: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := sftpPipe(t)
			defer cleanup()
			writeRemoteFile(t, client, "download.source", []byte(tt.remote))
			localPath := filepath.Join(t.TempDir(), "download.part")
			if err := os.WriteFile(localPath, []byte(tt.local), 0600); err != nil {
				t.Fatalf("write local destination: %v", err)
			}

			written, err := (&SFTPClient{client: client}).DownloadAt(context.Background(), "download.source", localPath, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DownloadAt error = %v, wantErr %v", err, tt.wantErr)
			}
			if written != tt.wantWritten {
				t.Fatalf("written = %d, want %d", written, tt.wantWritten)
			}
			content, readErr := os.ReadFile(localPath)
			if readErr != nil {
				t.Fatalf("read local destination: %v", readErr)
			}
			if string(content) != tt.wantContent {
				t.Fatalf("local content = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestSFTPClientReadAt(t *testing.T) {
	client, cleanup := sftpPipe(t)
	defer cleanup()
	sftpClient := &SFTPClient{client: client}
	writeRemoteFile(t, client, "range.txt", []byte("abcdefghij"))

	content, err := sftpClient.ReadAt(context.Background(), "range.txt", 3, 4)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if string(content) != "defg" {
		t.Fatalf("content = %q, want %q", content, "defg")
	}
	empty, err := sftpClient.ReadAt(context.Background(), "range.txt", 10, 0)
	if err != nil {
		t.Fatalf("ReadAt empty: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty read length = %d", len(empty))
	}
	if _, err := sftpClient.ReadAt(context.Background(), "range.txt", 8, 3); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("short ReadAt error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestSFTPClientRename(t *testing.T) {
	client, cleanup := sftpPipe(t)
	defer cleanup()
	sftpClient := &SFTPClient{client: client}
	writeRemoteFile(t, client, "source.part", []byte("new"))
	writeRemoteFile(t, client, "final.txt", []byte("old"))

	if err := sftpClient.Rename(context.Background(), "source.part", "final.txt"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if got := string(readRemoteFile(t, client, "final.txt")); got != "new" {
		t.Fatalf("renamed content = %q, want new", got)
	}
	if err := sftpClient.Rename(context.Background(), "missing.part", "other.txt"); err == nil {
		t.Fatal("rename of a missing source must fail")
	}
}

func TestSFTPResumePrimitivesPropagateCancellation(t *testing.T) {
	client, cleanup := sftpPipe(t)
	defer cleanup()
	sftpClient := &SFTPClient{client: client}
	writeRemoteFile(t, client, "source", []byte("content"))
	localPath := filepath.Join(t.TempDir(), "local")
	if err := os.WriteFile(localPath, []byte("content"), 0600); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	operations := []struct {
		name string
		run  func() error
	}{
		{name: "stat", run: func() error { _, err := sftpClient.Stat(ctx, "source"); return err }},
		{name: "upload", run: func() error { _, err := sftpClient.UploadAt(ctx, localPath, "upload", 0); return err }},
		{name: "upload range", run: func() error { _, err := sftpClient.UploadRangeAt(ctx, localPath, "upload", 0, 1); return err }},
		{name: "download", run: func() error { _, err := sftpClient.DownloadAt(ctx, "source", localPath, 0); return err }},
		{name: "download range", run: func() error { _, err := sftpClient.DownloadRangeAt(ctx, "source", localPath, 0, 1); return err }},
		{name: "read", run: func() error { _, err := sftpClient.ReadAt(ctx, "source", 0, 1); return err }},
		{name: "rename", run: func() error { return sftpClient.Rename(ctx, "source", "renamed") }},
		{name: "remove", run: func() error { return sftpClient.Remove(ctx, "source") }},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.run(); !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v, want context.Canceled", err)
			}
		})
	}
}

func TestSFTPClientRangeTransfersAndRemove(t *testing.T) {
	client, cleanup := sftpPipe(t)
	defer cleanup()
	sftpClient := &SFTPClient{client: client}
	localSource := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(localSource, []byte("abcdefghij"), 0600); err != nil {
		t.Fatalf("write local source: %v", err)
	}
	writeRemoteFile(t, client, "upload.part", []byte("abcXXXXX"))

	written, err := sftpClient.UploadRangeAt(context.Background(), localSource, "upload.part", 3, 4)
	if err != nil {
		t.Fatalf("UploadRangeAt: %v", err)
	}
	if written != 4 || string(readRemoteFile(t, client, "upload.part")) != "abcdefg" {
		t.Fatalf("range upload wrote %d bytes with content %q", written, readRemoteFile(t, client, "upload.part"))
	}
	if _, err := sftpClient.UploadRangeAt(context.Background(), localSource, "upload.part", 8, 3); err == nil {
		t.Fatal("upload range beyond source must fail")
	}

	localDest := filepath.Join(t.TempDir(), "download.part")
	if err := os.WriteFile(localDest, []byte("abcXXXXX"), 0600); err != nil {
		t.Fatalf("write local destination: %v", err)
	}
	written, err = sftpClient.DownloadRangeAt(context.Background(), "upload.part", localDest, 3, 4)
	if err != nil {
		t.Fatalf("DownloadRangeAt: %v", err)
	}
	content, err := os.ReadFile(localDest)
	if err != nil {
		t.Fatalf("read local destination: %v", err)
	}
	if written != 4 || string(content) != "abcdefg" {
		t.Fatalf("range download wrote %d bytes with content %q", written, content)
	}
	if _, err := sftpClient.DownloadRangeAt(context.Background(), "upload.part", localDest, 6, 2); err == nil {
		t.Fatal("download range beyond source must fail")
	}
	if err := sftpClient.Remove(context.Background(), "upload.part"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := client.Stat("upload.part"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed file stat error = %v, want os.ErrNotExist", err)
	}
}
