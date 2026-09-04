package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	opsfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/file"
)

func runRelay(ctx context.Context, args []string, out io.Writer) error {
	if ctx == nil {
		return fmt.Errorf("relay context is nil")
	}
	if len(args) == 0 {
		return fmt.Errorf("relay subcommand is required: serve or fetch")
	}
	switch args[0] {
	case "serve":
		return runRelayServe(ctx, args[1:], out)
	case "fetch":
		return runRelayFetch(ctx, args[1:], out)
	default:
		return fmt.Errorf("unknown relay subcommand %q", args[0])
	}
}

func runRelayServe(ctx context.Context, args []string, out io.Writer) (returnErr error) {
	flags := flag.NewFlagSet("relay serve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	filePath := flags.String("file", "", "single file to serve")
	listen := flags.String("listen", "127.0.0.1:0", "HTTPS listen address")
	advertiseHost := flags.String("advertise-host", "", "host targets use to reach the relay")
	ttl := flags.Duration("ttl", 5*time.Minute, "maximum relay lifetime")
	maxConcurrent := flags.Int("max-concurrent", 32, "maximum concurrent downloads")
	detach := flags.Bool("detach", false, "start a bounded background relay")
	workerInfo := flags.String("worker-info", "", "internal relay readiness file")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse relay serve flags: %w", err)
	}
	if *detach {
		return launchDetachedRelay(args, out)
	}
	server, err := opsfile.StartRelayHTTPServerWithOptions(*filePath, opsfile.RelayHTTPServerOptions{
		ListenAddress: *listen,
		AdvertiseHost: *advertiseHost,
		TTL:           *ttl,
		MaxConcurrent: *maxConcurrent,
	})
	if err != nil {
		return err
	}
	if *workerInfo != "" {
		encoded, encodeErr := json.Marshal(server.Info)
		if encodeErr != nil {
			return fmt.Errorf("encode relay worker info: %w", encodeErr)
		}
		if writeErr := os.WriteFile(*workerInfo, encoded, 0o600); writeErr != nil {
			return fmt.Errorf("write relay worker info: %w", writeErr)
		}
		defer func() {
			if removeErr := os.Remove(*workerInfo); removeErr != nil && !os.IsNotExist(removeErr) {
				returnErr = errors.Join(returnErr, fmt.Errorf("remove relay worker info: %w", removeErr))
			}
		}()
	}
	if err := json.NewEncoder(out).Encode(server.Info); err != nil {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		warnings := server.Stop(stopCtx)
		if len(warnings) > 0 {
			return fmt.Errorf("write relay session: %w; cleanup warnings: %v", err, warnings)
		}
		return fmt.Errorf("write relay session: %w", err)
	}
	timer := time.NewTimer(time.Until(server.Info.ExpiresAt))
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if warnings := server.Stop(stopCtx); len(warnings) > 0 {
		return fmt.Errorf("relay cleanup warnings: %v", warnings)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func launchDetachedRelay(args []string, out io.Writer) (returnErr error) {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve runner executable: %w", err)
	}
	infoFile, err := os.CreateTemp("", ".opslang-relay-info-*")
	if err != nil {
		return fmt.Errorf("create relay readiness file: %w", err)
	}
	infoPath := infoFile.Name()
	if err := infoFile.Close(); err != nil {
		return fmt.Errorf("close relay readiness file: %w", err)
	}
	if err := os.Remove(infoPath); err != nil {
		return fmt.Errorf("prepare relay readiness file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(infoPath); removeErr != nil && !os.IsNotExist(removeErr) {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove relay readiness file: %w", removeErr))
		}
	}()
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open null device: %w", err)
	}
	defer devNull.Close()
	childArgs := []string{executable, "relay", "serve"}
	childArgs = append(childArgs, args...)
	childArgs = append(childArgs, "--detach=false", "--worker-info", infoPath)
	process, err := os.StartProcess(executable, childArgs, &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{devNull, devNull, devNull},
	})
	if err != nil {
		return fmt.Errorf("start detached relay: %w", err)
	}
	if err := process.Release(); err != nil {
		return fmt.Errorf("release detached relay process: %w", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		encoded, readErr := os.ReadFile(infoPath)
		if readErr == nil {
			var info opsfile.RelayHTTPInfo
			if json.Unmarshal(encoded, &info) == nil && info.URL != "" {
				if err := json.NewEncoder(out).Encode(info); err != nil {
					return fmt.Errorf("write detached relay info: %w", err)
				}
				return nil
			}
		} else if !os.IsNotExist(readErr) {
			return fmt.Errorf("read relay readiness file: %w", readErr)
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("detached relay did not become ready within 5s")
}

func runRelayFetch(ctx context.Context, args []string, out io.Writer) error {
	flags := flag.NewFlagSet("relay fetch", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	url := flags.String("url", "", "relay HTTPS URL")
	token := flags.String("token", "", "relay bearer token")
	fingerprint := flags.String("fingerprint", "", "relay certificate SHA-256 fingerprint")
	checksum := flags.String("sha256", "", "expected file SHA-256")
	size := flags.Int64("size", -1, "expected file size")
	wireChecksum := flags.String("wire-sha256", "", "transferred object SHA-256")
	wireSize := flags.Int64("wire-size", -1, "transferred object size")
	decompress := flags.Bool("decompress", false, "decompress after transfer")
	destination := flags.String("dest", "", "final destination path")
	retention := flags.Duration("part-retention", 24*time.Hour, "partial file retention")
	timeout := flags.Duration("timeout", 60*time.Second, "HTTP request timeout")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse relay fetch flags: %w", err)
	}
	outcome, err := opsfile.RelayFetch(ctx, opsfile.RelayFetchOptions{
		URL:             *url,
		Token:           *token,
		CertFingerprint: *fingerprint,
		SHA256:          *checksum,
		Size:            *size,
		WireSHA256:      *wireChecksum,
		WireSize:        *wireSize,
		Decompress:      *decompress,
		Dest:            *destination,
		PartRetention:   *retention,
		Timeout:         *timeout,
	})
	if err != nil {
		return err
	}
	if err := json.NewEncoder(out).Encode(outcome); err != nil {
		return fmt.Errorf("write relay fetch result: %w", err)
	}
	return nil
}
