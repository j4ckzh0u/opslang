package file

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func gzipFile(source string) (string, error) {
	input, err := os.Open(source)
	if err != nil {
		return "", fmt.Errorf("open compression source: %w", err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return "", fmt.Errorf("stat compression source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("compression source %s is not a regular file", source)
	}
	output, err := os.CreateTemp(filepath.Dir(source), ".opslang-compressed-*")
	if err != nil {
		return "", fmt.Errorf("create compressed staging file: %w", err)
	}
	path := output.Name()
	writer := gzip.NewWriter(output)
	if _, err := io.Copy(writer, input); err != nil {
		_ = writer.Close()
		_ = output.Close()
		return path, fmt.Errorf("compress file: %w", err)
	}
	if err := writer.Close(); err != nil {
		_ = output.Close()
		return path, fmt.Errorf("finish compression: %w", err)
	}
	if err := output.Close(); err != nil {
		return path, fmt.Errorf("close compressed staging file: %w", err)
	}
	return path, nil
}

func gunzipFile(source, destination string) (returnErr error) {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open compressed file: %w", err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil && returnErr == nil {
			returnErr = fmt.Errorf("close compressed file: %w", closeErr)
		}
	}()
	reader, err := gzip.NewReader(input)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	output, err := os.Create(destination)
	if err != nil {
		_ = reader.Close()
		return fmt.Errorf("create decompressed staging file: %w", err)
	}
	if _, err := io.Copy(output, reader); err != nil {
		returnErr = fmt.Errorf("decompress file: %w", err)
	}
	if err := reader.Close(); err != nil && returnErr == nil {
		returnErr = fmt.Errorf("close gzip stream: %w", err)
	}
	if err := output.Close(); err != nil && returnErr == nil {
		returnErr = fmt.Errorf("close decompressed staging file: %w", err)
	}
	return returnErr
}
