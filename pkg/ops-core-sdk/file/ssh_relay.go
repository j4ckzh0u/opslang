package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/j4ckzh0u/opslang/internal/sshx"
)

// SSHRelayGroup uploads one seed and asks peers to pull it through HTTPS.
func SSHRelayGroup(ctx context.Context, source string, relay DistributeTarget, targets []DistributeTarget, opts DistributeOptions) (map[string]HostDistributeResult, error) {
	transferSource := source
	if opts.Compress {
		compressed, compressErr := gzipFile(source)
		if compressErr != nil {
			return nil, compressErr
		}
		defer os.Remove(compressed)
		transferSource = compressed
	}
	remotePath := targetRemotePath(source, relay)
	endpoint := formatEndpoint(effectiveTargetUser(relay), relay.Host, effectiveTargetPort(relay), remotePath)
	seed, err := SSHResumeUpload(ctx, transferSource, endpoint, opts.PartRetention)
	if err != nil {
		return nil, fmt.Errorf("upload relay seed: %w", err)
	}
	if opts.Compress {
		info, infoErr := os.Stat(source)
		if infoErr != nil {
			return nil, fmt.Errorf("stat relay source: %w", infoErr)
		}
		checksum, checksumErr := computeFileChecksum(source)
		if checksumErr != nil {
			return nil, fmt.Errorf("checksum relay source: %w", checksumErr)
		}
		seed.Size, seed.Checksum = info.Size(), checksum
	}
	if opts.Mode != "" {
		if err := SSHChmod(ctx, endpoint, opts.Mode); err != nil {
			return nil, fmt.Errorf("chmod relay seed: %w", err)
		}
	}
	runnerPath := os.Getenv("OPSLANG_REMOTE_RUNNER")
	if runnerPath == "" {
		runnerPath = "ops-runner"
	}
	ttl := opts.Timeout
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	var session RelayHTTPInfo
	err = withSSHClient(ctx, endpoint, func(client *sshx.Client) error {
		command := shellCommand(runnerPath, "relay", "serve", "--detach", "--file", remotePath, "--listen", "0.0.0.0:0", "--advertise-host", relay.Host, "--ttl", ttl.String(), "--max-concurrent", strconv.Itoa(max(1, opts.Parallel)))
		result, execErr := client.Exec(ctx, command)
		if execErr != nil {
			return execErr
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("relay serve exited %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		if err := json.Unmarshal([]byte(result.Stdout), &session); err != nil {
			return fmt.Errorf("decode relay session: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("start relay service: %w", err)
	}
	outcomes := map[string]HostDistributeResult{
		relayTargetIdentity(relay): transferOutcomeResult(relay.Host, seed, "relay_seed", ""),
	}
	parallel := opts.Parallel
	if parallel <= 0 {
		parallel = 5
	}
	semaphore := make(chan struct{}, parallel)
	var mu sync.Mutex
	var wait sync.WaitGroup
	for _, target := range targets {
		wait.Add(1)
		go func(target DistributeTarget) {
			defer wait.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			outcome, fetchErr := sshRelayFetch(ctx, runnerPath, session, target, source, opts)
			result := transferOutcomeResult(target.Host, outcome, "relay:"+relay.Host, "")
			if fetchErr != nil {
				result.Status = "failed"
				result.Error = fetchErr.Error()
			}
			mu.Lock()
			outcomes[relayTargetIdentity(target)] = result
			mu.Unlock()
		}(target)
	}
	wait.Wait()
	return outcomes, nil
}

func sshRelayFetch(ctx context.Context, runnerPath string, session RelayHTTPInfo, target DistributeTarget, source string, opts DistributeOptions) (TransferOutcome, error) {
	endpoint := formatEndpoint(effectiveTargetUser(target), target.Host, effectiveTargetPort(target), targetRemotePath(source, target))
	var outcome TransferOutcome
	finalInfo, err := os.Stat(source)
	if err != nil {
		return outcome, fmt.Errorf("stat relay source: %w", err)
	}
	finalChecksum, err := computeFileChecksum(source)
	if err != nil {
		return outcome, fmt.Errorf("checksum relay source: %w", err)
	}
	err = withSSHClient(ctx, endpoint, func(client *sshx.Client) error {
		commandArgs := []string{runnerPath, "relay", "fetch", "--url", session.URL, "--token", session.Token, "--fingerprint", session.CertFingerprint, "--sha256", finalChecksum, "--size", strconv.FormatInt(finalInfo.Size(), 10), "--wire-sha256", session.SHA256, "--wire-size", strconv.FormatInt(session.Size, 10), "--dest", targetRemotePath(source, target), "--part-retention", effectivePartRetention(opts.PartRetention).String()}
		if opts.Compress {
			commandArgs = append(commandArgs, "--decompress")
		}
		command := shellCommand(commandArgs...)
		result, execErr := client.Exec(ctx, command)
		if execErr != nil {
			return execErr
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("relay fetch exited %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		return json.Unmarshal([]byte(result.Stdout), &outcome)
	})
	return outcome, err
}

func effectiveTargetUser(target DistributeTarget) string {
	if target.User == "" {
		return "root"
	}
	return target.User
}

func effectiveTargetPort(target DistributeTarget) int {
	if target.Port == 0 {
		return 22
	}
	return target.Port
}

func shellCommand(arguments ...string) string {
	quoted := make([]string, len(arguments))
	for index, argument := range arguments {
		quoted[index] = "'" + strings.ReplaceAll(argument, "'", "'\"'\"'") + "'"
	}
	return strings.Join(quoted, " ")
}

func transferOutcomeResult(host string, outcome TransferOutcome, source, errorMessage string) HostDistributeResult {
	status := outcome.Status
	if status == "" {
		status = "success"
	}
	return HostDistributeResult{Host: host, Status: status, Changed: outcome.Changed, Checksum: outcome.Checksum, Size: outcome.Size, TransferSource: source, ResumedBytes: outcome.ResumedBytes, TransferredBytes: outcome.TransferredBytes, Warnings: outcome.Warnings, Error: errorMessage}
}
