package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type relayScaleCluster struct {
	*scaleCluster

	mu                  sync.Mutex
	sourceSize          int64
	sourceChecksum      string
	partialOffsets      map[string]int64
	invalidOffsets      map[string]bool
	corruptPrefixes     map[string]bool
	corruptTransfers    map[string]bool
	candidateAttempts   map[string]int
	failedCandidate     bool
	seedEgressBytes     int64
	rangeBytes          int64
	relayEgressBytes    int64
	directFallbacks     int
	invalidOffsetResets int
	corruptPrefixResets int
}

func newRelayScaleCluster(t *testing.T, hosts []string, fileSize int64, checksum string) *relayScaleCluster {
	t.Helper()
	base := newScaleCluster(t, len(hosts), int(fileSize/1024), 0, 0)
	cluster := &relayScaleCluster{
		scaleCluster:      base,
		sourceSize:        fileSize,
		sourceChecksum:    checksum,
		partialOffsets:    make(map[string]int64, len(hosts)),
		invalidOffsets:    make(map[string]bool),
		corruptPrefixes:   make(map[string]bool),
		corruptTransfers:  make(map[string]bool),
		candidateAttempts: make(map[string]int),
	}
	cluster.hosts = append(cluster.hosts[:0], hosts...)
	cluster.remote = make(map[string]map[string]*scaleRemoteFile, len(hosts))
	cluster.hostWireBytes = make(map[string]int64, len(hosts))
	for index, host := range hosts {
		cluster.remote[host] = make(map[string]*scaleRemoteFile, 1)
		if index%2 == 0 {
			cluster.partialOffsets[host] = fileSize / 2
		}
		if index%1000 == 17 {
			cluster.corruptTransfers[host] = true
		}
		if index%1000 == 29 {
			cluster.partialOffsets[host] = fileSize + 1
			cluster.invalidOffsets[host] = true
		}
		if index%1000 == 41 {
			cluster.partialOffsets[host] = fileSize / 2
			cluster.corruptPrefixes[host] = true
		}
	}
	return cluster
}

func (sc *relayScaleCluster) relayGroup(_ context.Context, _ string, relay DistributeTarget, peers []DistributeTarget, _ DistributeOptions) (map[string]HostDistributeResult, error) {
	sc.mu.Lock()
	sc.candidateAttempts[relay.Host]++
	if !sc.failedCandidate {
		sc.failedCandidate = true
		sc.seedEgressBytes += sc.sourceSize * 2 / 5
		sc.mu.Unlock()
		return nil, fmt.Errorf("injected relay candidate failure")
	}
	sc.seedEgressBytes += sc.sourceSize
	sc.mu.Unlock()

	outcomes := make(map[string]HostDistributeResult, len(peers)+1)
	sc.store(relay.Host, targetRemotePath("payload.bin", relay), sc.sourceSize, sc.sourceChecksum)
	outcomes[relayTargetIdentity(relay)] = HostDistributeResult{
		Host:             relay.Host,
		Status:           "success",
		Changed:          true,
		Checksum:         sc.sourceChecksum,
		Size:             sc.sourceSize,
		TransferSource:   "relay_seed",
		TransferredBytes: sc.sourceSize,
	}

	for _, peer := range peers {
		offset, warnings := sc.validatedOffset(peer.Host)
		transferred := sc.sourceSize - offset
		sc.mu.Lock()
		sc.rangeBytes += transferred
		sc.relayEgressBytes += transferred
		corrupt := sc.corruptTransfers[peer.Host]
		sc.mu.Unlock()
		if corrupt {
			sc.store(peer.Host, targetRemotePath("payload.bin", peer), sc.sourceSize, "corrupted:"+sc.sourceChecksum)
			outcomes[relayTargetIdentity(peer)] = HostDistributeResult{
				Host: peer.Host, Status: "failed", Checksum: sc.sourceChecksum,
				Size: sc.sourceSize, TransferSource: "relay:" + relay.Host,
				ResumedBytes: offset, TransferredBytes: transferred,
				Warnings: warnings, Error: "checksum mismatch (injected)",
			}
			continue
		}
		sc.store(peer.Host, targetRemotePath("payload.bin", peer), sc.sourceSize, sc.sourceChecksum)
		outcomes[relayTargetIdentity(peer)] = HostDistributeResult{
			Host: peer.Host, Status: "success", Changed: true,
			Checksum: sc.sourceChecksum, Size: sc.sourceSize,
			TransferSource: "relay:" + relay.Host, ResumedBytes: offset,
			TransferredBytes: transferred, Warnings: warnings,
		}
	}
	return outcomes, nil
}

func (sc *relayScaleCluster) validatedOffset(host string) (int64, []string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	offset := sc.partialOffsets[host]
	if sc.invalidOffsets[host] {
		sc.invalidOffsetResets++
		return 0, []string{"invalid resume offset reset to zero"}
	}
	if sc.corruptPrefixes[host] {
		sc.corruptPrefixResets++
		return 0, []string{"resume prefix mismatch reset to zero"}
	}
	return offset, nil
}

func (sc *relayScaleCluster) directUpload(ctx context.Context, source, endpoint string) error {
	if err := sc.upload(ctx, source, endpoint); err != nil {
		return err
	}
	sc.mu.Lock()
	sc.directFallbacks++
	sc.mu.Unlock()
	return nil
}

func runScaleRelayDistribute(t *testing.T, n int) {
	t.Helper()
	fileKB := scaleEnvInt(t, "OPS_SCALE_FILE_KB", 64)
	dir := t.TempDir()
	source := filepath.Join(dir, "payload.bin")
	payload := make([]byte, fileKB*1024)
	for index := range payload {
		payload[index] = byte(index % 251)
	}
	if err := os.WriteFile(source, payload, 0o600); err != nil {
		t.Fatalf("write relay scale source: %v", err)
	}
	checksum, err := computeFileChecksum(source)
	if err != nil {
		t.Fatalf("checksum relay scale source: %v", err)
	}

	hosts := make([]string, n)
	targets := make([]DistributeTarget, n)
	for index := range targets {
		host := fmt.Sprintf("relay-sim-%05d.test", index)
		hosts[index] = host
		targets[index] = DistributeTarget{
			Host: host, User: "root", Port: 22,
			Dest: "/opt/opslang/payload.bin", RelayGroup: "scale-zone",
		}
	}
	cluster := newRelayScaleCluster(t, hosts, int64(len(payload)), checksum)
	opts := DistributeOptions{
		Checksum: true, Mode: "0644", Parallel: 64, Retries: 3,
		Timeout: 30 * time.Second, Relay: true, RelayThreshold: 20,
		RelayMaxTargets: 100,
	}
	plan, err := BuildRelayPlan(targets, opts)
	if err != nil {
		t.Fatalf("build relay scale plan: %v", err)
	}
	directPlanTargets := 0
	relayPlanGroups := 0
	for _, group := range plan.Groups {
		directPlanTargets += len(group.Direct)
		if group.Relay != nil {
			relayPlanGroups++
		}
	}

	savedRelay, savedVerify, savedChmod := DefaultRelayGroupFunc, DefaultVerifyFunc, DefaultChmodFunc
	DefaultRelayGroupFunc = cluster.relayGroup
	DefaultVerifyFunc = cluster.verify
	DefaultChmodFunc = cluster.chmod
	t.Cleanup(func() {
		DefaultRelayGroupFunc, DefaultVerifyFunc, DefaultChmodFunc = savedRelay, savedVerify, savedChmod
	})

	result, err := DistributeWith(source, targets, opts, cluster.directUpload)
	if err != nil {
		t.Fatalf("relay scale distribute: %v", err)
	}
	if result.Total != n || len(result.Results) != n {
		t.Fatalf("result conservation: total=%d results=%d want=%d", result.Total, len(result.Results), n)
	}
	if result.Succeeded+result.Skipped+result.Failed != n {
		t.Fatalf("status conservation: success=%d skipped=%d failed=%d want=%d", result.Succeeded, result.Skipped, result.Failed, n)
	}
	requireSuccessRate(t, "relay distribute", result.Succeeded, result.Total)

	seen := make(map[string]bool, n)
	var resumedResults int
	var candidateWarnings int
	for _, hostResult := range result.Results {
		if seen[hostResult.Host] {
			t.Errorf("duplicate relay result for %s", hostResult.Host)
		}
		seen[hostResult.Host] = true
		if hostResult.Status != "success" || hostResult.Checksum != checksum || hostResult.Size != int64(len(payload)) {
			t.Errorf("invalid result for %s: %+v", hostResult.Host, hostResult)
		}
		if hostResult.ResumedBytes > 0 {
			resumedResults++
		}
		for _, warning := range hostResult.Warnings {
			if strings.Contains(warning, "candidate") {
				candidateWarnings++
			}
		}
		record := cluster.lookup(hostResult.Host, "/opt/opslang/payload.bin")
		if record == nil || record.checksum != checksum || record.size != int64(len(payload)) {
			t.Errorf("invalid remote state for %s: %+v", hostResult.Host, record)
		}
	}
	if len(seen) != n {
		t.Errorf("unique result hosts=%d want=%d", len(seen), n)
	}
	if resumedResults == 0 || cluster.rangeBytes == 0 {
		t.Errorf("resume path was not exercised: results=%d rangeBytes=%d", resumedResults, cluster.rangeBytes)
	}
	if cluster.invalidOffsetResets == 0 || cluster.corruptPrefixResets == 0 {
		t.Errorf("resume faults not exercised: invalid=%d corrupt=%d", cluster.invalidOffsetResets, cluster.corruptPrefixResets)
	}
	wantDirect := len(cluster.corruptTransfers) + directPlanTargets
	if cluster.directFallbacks != wantDirect {
		t.Errorf("direct fallbacks=%d want=%d", cluster.directFallbacks, wantDirect)
	}
	if candidateWarnings == 0 || !cluster.failedCandidate {
		t.Errorf("relay candidate failure was not observable: warnings=%d", candidateWarnings)
	}

	groups := len(plan.Groups)
	controllerEgress := cluster.seedEgressBytes + cluster.egressBytes.Load()
	controllerMax := int64(relayPlanGroups+1+wantDirect) * int64(len(payload))
	if controllerEgress > controllerMax {
		t.Errorf("controller egress=%d exceeds relay bound=%d", controllerEgress, controllerMax)
	}
	if controllerEgress >= int64(n)*int64(len(payload))/10 {
		t.Errorf("controller egress=%d did not stay below 10%% direct fan-out=%d", controllerEgress, int64(n)*int64(len(payload)))
	}
	if cluster.relayEgressBytes > int64(n)*int64(len(payload)) {
		t.Errorf("relay egress=%d exceeds one full copy per target=%d", cluster.relayEgressBytes, int64(n)*int64(len(payload)))
	}
	for host, attempts := range cluster.candidateAttempts {
		if attempts > 1 {
			t.Errorf("candidate %s attempts=%d want <=1", host, attempts)
		}
	}

	t.Logf("relay scale: hosts=%d groups=%d success=%d controller=%dB relay=%dB range=%dB fallbacks=%d",
		n, groups, result.Succeeded, controllerEgress, cluster.relayEgressBytes, cluster.rangeBytes, cluster.directFallbacks)
}

func TestScaleRelayDistribute10000Hosts(t *testing.T) {
	if testing.Short() {
		t.Skip("relay scale simulation skipped in -short mode; run `make scale-test` for the full tier")
	}
	runScaleRelayDistribute(t, scaleEnvInt(t, "OPS_SCALE_N", 10000))
}

func TestScaleRelayDistributeCI(t *testing.T) {
	runScaleRelayDistribute(t, scaleEnvInt(t, "OPS_SCALE_CI_N", 1000))
}
