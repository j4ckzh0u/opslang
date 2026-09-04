package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/internal/sshx"
)

// SSHCollectRelayGroup implements fan-in through a relay host. The source
// host owns the HTTPS session; the relay only receives the short-lived session
// metadata and writes a temporary copy for the controller to download.
func SSHCollectRelayGroup(ctx context.Context, source string, relay CollectTarget, targets []CollectTarget, opts CollectOptions) (map[string]HostCollectResult, error) {
	if opts.Compress {
		return nil, fmt.Errorf("relay collect does not support compressed sessions")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("relay collect requires at least one target")
	}
	runnerPath := os.Getenv("OPSLANG_REMOTE_RUNNER")
	if runnerPath == "" {
		runnerPath = "ops-runner"
	}
	ttl := opts.Timeout
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	outcomes := make(map[string]HostCollectResult, len(targets)+1)
	allTargets := append([]CollectTarget{relay}, targets...)
	for _, target := range allTargets {
		hostResult := collectOneThroughRelay(ctx, runnerPath, source, relay, target, ttl, opts)
		outcomes[collectTargetIdentity(target)] = hostResult
	}
	return outcomes, nil
}

func collectOneThroughRelay(ctx context.Context, runnerPath, fallbackSource string, relay, target CollectTarget, ttl time.Duration, opts CollectOptions) HostCollectResult {
	localDest := collectDestination(opts.DestDir, target, fallbackSource)
	hostResult := HostCollectResult{Host: target.Host, Source: target.Source, Dest: localDest, TransferSource: "relay_collect"}
	remoteSource := effectiveCollectSource(target, fallbackSource)
	if strings.TrimSpace(remoteSource) == "" {
		hostResult.Status, hostResult.Error = "failed", "relay collect source path is empty"
		return hostResult
	}
	sourceTarget := target
	sourceTarget.Source = remoteSource
	session, err := startSourceRelay(ctx, runnerPath, sourceTarget, relay.Host, ttl, opts.Parallel)
	if err != nil {
		hostResult.Status, hostResult.Error = "failed", fmt.Sprintf("start source relay: %v", err)
		return hostResult
	}
	relayPath, err := randomRelayPath("collect")
	if err != nil {
		hostResult.Status, hostResult.Error = "failed", err.Error()
		return hostResult
	}
	relayEndpoint := formatEndpoint(effectiveCollectUser(relay), relay.Host, effectiveCollectPort(relay), relayPath)
	if err := fetchSourceOnRelay(ctx, runnerPath, relayEndpoint, relayPath, session, opts); err != nil {
		hostResult.Status, hostResult.Error = "failed", err.Error()
		return hostResult
	}
	defer cleanupRemoteFile(context.Background(), relayEndpoint, relayPath)
	var outcome TransferOutcome
	if opts.Resume {
		outcome, err = SSHResumeDownload(ctx, relayEndpoint, localDest, opts.PartRetention)
	} else {
		err = SSHCollectDownload(ctx, relayEndpoint, localDest)
		outcome.TransferSource = "relay_collect"
		if err == nil {
			info, statErr := os.Stat(localDest)
			if statErr != nil {
				err = statErr
			} else {
				outcome.Status, outcome.Changed, outcome.Size = "success", true, info.Size()
				outcome.Checksum, err = computeFileChecksum(localDest)
			}
		}
	}
	if err != nil {
		hostResult.Status, hostResult.Error = "failed", err.Error()
		return hostResult
	}
	hostResult.Status = outcome.Status
	if hostResult.Status == "" {
		hostResult.Status = "success"
	}
	hostResult.Checksum, hostResult.Size = outcome.Checksum, outcome.Size
	hostResult.ResumedBytes, hostResult.TransferredBytes = outcome.ResumedBytes, outcome.TransferredBytes
	hostResult.Warnings = outcome.Warnings
	return hostResult
}

func startSourceRelay(ctx context.Context, runnerPath string, source CollectTarget, relayHost string, ttl time.Duration, parallel int) (RelayHTTPInfo, error) {
	endpoint := formatEndpoint(effectiveCollectUser(source), source.Host, effectiveCollectPort(source), effectiveCollectSource(source, ""))
	var session RelayHTTPInfo
	err := withSSHClient(ctx, endpoint, func(client *sshx.Client) error {
		command := shellCommand(runnerPath, "relay", "serve", "--detach", "--file", effectiveCollectSource(source, ""), "--listen", "0.0.0.0:0", "--advertise-host", relayHost, "--ttl", ttl.String(), "--max-concurrent", strconv.Itoa(max(1, parallel)))
		result, execErr := client.Exec(ctx, command)
		if execErr != nil {
			return execErr
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("relay serve exited %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		if err := json.Unmarshal([]byte(result.Stdout), &session); err != nil {
			return fmt.Errorf("decode source relay session: %w", err)
		}
		return nil
	})
	return session, err
}

func fetchSourceOnRelay(ctx context.Context, runnerPath, relayEndpoint, relayPath string, session RelayHTTPInfo, opts CollectOptions) error {
	return withSSHClient(ctx, relayEndpoint, func(client *sshx.Client) error {
		command := shellCommand(runnerPath, "relay", "fetch", "--url", session.URL, "--token", session.Token, "--fingerprint", session.CertFingerprint, "--sha256", session.SHA256, "--size", strconv.FormatInt(session.Size, 10), "--wire-sha256", session.SHA256, "--wire-size", strconv.FormatInt(session.Size, 10), "--dest", relayPath, "--part-retention", effectivePartRetention(opts.PartRetention).String())
		result, err := client.Exec(ctx, command)
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("relay fetch exited %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
		}
		return nil
	})
}

func cleanupRemoteFile(ctx context.Context, endpoint, path string) {
	_ = withSSHClient(ctx, endpoint, func(client *sshx.Client) error {
		sftp, err := client.NewSFTPClient()
		if err != nil {
			return err
		}
		defer sftp.Close()
		return sftp.Remove(ctx, path)
	})
}

func randomRelayPath(prefix string) (string, error) {
	value, err := randomHex(12)
	if err != nil {
		return "", err
	}
	return filepath.Join(os.TempDir(), ".opslang-"+prefix+"-"+value), nil
}

func effectiveCollectSource(target CollectTarget, fallback string) string {
	if target.Source != "" {
		return target.Source
	}
	return fallback
}

func effectiveCollectUser(target CollectTarget) string {
	if target.User == "" {
		return "root"
	}
	return target.User
}

func effectiveCollectPort(target CollectTarget) int {
	if target.Port == 0 {
		return 22
	}
	return target.Port
}
