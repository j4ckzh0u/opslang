// Package file - scale_test.go: 10k-host distribute/collect simulation.
//
// These tests stress the REAL orchestration logic in DistributeWith /
// CollectWith (fan-out scheduling, concurrency semaphore, per-attempt
// timeout, retry with backoff, endpoint formatting, result aggregation,
// per-host archiving, post-download checksum/size recording) against a
// simulated fleet of N hosts.
//
// What is mocked (the transport seam, mirroring the real SFTP layer in
// internal/sshx):
//
//		TransferFunc          - "upload":   streams the real local source file
//		                                      (read + SHA-256, like a real
//	                                     controller pushing bytes), records
//	                                     {size, checksum} on the virtual host.
//		VerifyFunc            - "checksum": compares against what the virtual
//		                                      host actually received.
//		ChmodFunc             - "chmod":    records the mode remotely.
//		CollectDownloadFunc   - "download": MkdirAll + writes the virtual host's
//		                                      remote content to dst (same
//		                                      parent-dir semantics as
//		                                      sshx.SFTPClient.Download).
//
// What is NOT mocked: everything above the transport seam. The simulator
// also counts wire bytes at the transport boundary so the tests can assert
// the controller-side bandwidth profile, and injects small-probability
// transient failures (hard errors and silent corruption caught by the
// verify hook) so the retry path is exercised at scale.
//
// Scale knobs (env):
//
//	OPS_SCALE_N          host count            (default 10000; CI runs this)
//	OPS_SCALE_FILE_KB    per-file size in KB   (default 64)
//	OPS_SCALE_FAIL_RATE  per-attempt failure   (default 0.001 = 0.1%)
//	OPS_SCALE_LATENCY_MS per-transfer latency  (default 2)
//
// `go test -short` skips the big tiers; TestScale*_CI always run at 1000
// hosts as the CI gate. `make scale-test` runs the full 10k tiers.
package file

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ─── scale configuration ────────────────────────────────────────────

func scaleEnvInt(t *testing.T, key string, def int) int {
	t.Helper()
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			t.Fatalf("%s=%q: want a positive integer", key, v)
		}
		return n
	}
	return def
}

func scaleEnvFloat(t *testing.T, key string, def float64) float64 {
	t.Helper()
	if v := os.Getenv(key); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 || f >= 1 {
			t.Fatalf("%s=%q: want a float in [0,1)", key, v)
		}
		return f
	}
	return def
}

// ─── simulated fleet ────────────────────────────────────────────────

// scaleRemoteFile is what a virtual host ended up storing on its disk.
type scaleRemoteFile struct {
	size     int64
	checksum string // sha256 of bytes actually received (may be corrupted)
	mode     string // last chmod mode applied, e.g. "0644"
}

// scaleCluster simulates N hosts plus the wire between controller and
// fleet. It is the transport mock: DistributeWith/CollectWith run their
// real code against it.
type scaleCluster struct {
	mu      sync.Mutex
	hosts   []string
	remote  map[string]map[string]*scaleRemoteFile // host -> path -> file
	attempt map[string]int                         // host -> transfer call count (failure salt)

	failRate float64
	latency  time.Duration
	fileSize int64

	// Guaranteed quota of injected first-attempt failures: the k hosts
	// with the lowest fnv(host) rank fail their first transfer no matter
	// the fleet size, so even small tiers exercise the retry path
	// deterministically. k = max(1, round(N*failRate)).
	firstFail  map[string]bool // host -> first attempt fails
	corruptSet map[string]bool // ... and the failure is silent corruption

	// seedChecksum[host] = sha256 of the host's pre-seeded remote file
	// (unique per host so mis-attributed archives are caught).
	seedChecksum map[string]string

	egressBytes   atomic.Int64 // controller -> fleet (distribute)
	ingressBytes  atomic.Int64 // fleet -> controller (collect)
	hostWireBytes map[string]int64

	hardFails   atomic.Int64 // injected transport errors
	corruptions atomic.Int64 // silent corruptions caught by verify
}

func newScaleCluster(t *testing.T, n int, fileSizeKB int, failRate float64, latency time.Duration) *scaleCluster {
	t.Helper()
	sc := &scaleCluster{
		remote:        make(map[string]map[string]*scaleRemoteFile, n),
		attempt:       make(map[string]int, n),
		hosts:         make([]string, 0, n),
		failRate:      failRate,
		latency:       latency,
		fileSize:      int64(fileSizeKB) * 1024,
		firstFail:     make(map[string]bool),
		corruptSet:    make(map[string]bool),
		seedChecksum:  make(map[string]string, n),
		hostWireBytes: make(map[string]int64, n),
	}
	for i := 0; i < n; i++ {
		host := fmt.Sprintf("sim-%05d.test", i)
		sc.hosts = append(sc.hosts, host)
		sc.remote[host] = make(map[string]*scaleRemoteFile, 1)
		sc.seedChecksum[host] = sc.hostSeedChecksum(host)
	}

	// Deterministic failure quota: rank hosts by fnv(host) and pick the
	// first k to fail their opening transfer (half hard error, half
	// corruption). Unpredictable to the orchestration, reproducible in CI.
	if failRate > 0 {
		type ranked struct {
			v    uint64
			host string
		}
		ranks := make([]ranked, 0, n)
		for _, host := range sc.hosts {
			h := fnv.New64a()
			h.Write([]byte(host))
			ranks = append(ranks, ranked{h.Sum64(), host})
		}
		sort.Slice(ranks, func(i, j int) bool { return ranks[i].v < ranks[j].v })
		k := int(math.Round(float64(n) * failRate))
		if k == 0 {
			k = 1
		}
		if k > n {
			k = n
		}
		for _, r := range ranks[:k] {
			sc.firstFail[r.host] = true
			sc.corruptSet[r.host] = r.v&(1<<63) == 0 // deterministic split
		}
	}
	return sc
}

// hostBlock deterministically derives the host's 4KB content block.
func (sc *scaleCluster) hostBlock(host string) []byte {
	h := fnv.New64a()
	h.Write([]byte(host))
	rng := rand.New(rand.NewSource(int64(h.Sum64())))
	block := make([]byte, 4096)
	rng.Read(block)
	return block
}

// hostSeedChecksum precomputes the checksum of the host's remote file.
func (sc *scaleCluster) hostSeedChecksum(host string) string {
	block := sc.hostBlock(host)
	digest := sha256.New()
	remaining := sc.fileSize
	for remaining > 0 {
		n := int64(len(block))
		if remaining < n {
			n = remaining
		}
		digest.Write(block[:n])
		remaining -= n
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// writeHostContent streams the host's remote file content to w, exactly
// like reading from a remote SFTP file. Returns bytes written.
func (sc *scaleCluster) writeHostContent(host string, w io.Writer) (int64, error) {
	block := sc.hostBlock(host)
	var total int64
	remaining := sc.fileSize
	for remaining > 0 {
		n := int64(len(block))
		if remaining < n {
			n = remaining
		}
		if _, err := w.Write(block[:n]); err != nil {
			return total, err
		}
		total += n
		remaining -= n
	}
	return total, nil
}

// injectFailure decides, deterministically but unpredictably for the
// orchestration, whether a given attempt fails and how. The first attempt
// of quota-selected hosts always fails (guaranteeing the retry path runs
// at any fleet size); retries fail with per-attempt probability failRate.
// corrupt=true means the transfer reports success but delivered damaged
// bytes - only the verify hook can catch it.
func (sc *scaleCluster) injectFailure(host string) (fail, corrupt bool) {
	if sc.failRate <= 0 {
		return false, false
	}
	sc.mu.Lock()
	attemptNo := sc.attempt[host]
	sc.attempt[host] = attemptNo + 1
	sc.mu.Unlock()

	if attemptNo == 0 && sc.firstFail[host] {
		return true, sc.corruptSet[host]
	}

	h := fnv.New64a()
	h.Write([]byte(host))
	var salt [8]byte
	binary.LittleEndian.PutUint64(salt[:], uint64(attemptNo))
	h.Write(salt[:])
	v := float64(h.Sum64()%1_000_000) / 1_000_000
	if v >= sc.failRate {
		return false, false
	}
	// Same hash, top bit: pick hard error vs silent corruption.
	return true, h.Sum64()&(1<<63) == 0
}

// sleep simulates network latency, honouring the attempt deadline.
func (sc *scaleCluster) sleep(ctx context.Context) error {
	if sc.latency <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sc.latency):
		return nil
	}
}

func (sc *scaleCluster) store(host, path string, size int64, checksum string) {
	sc.mu.Lock()
	sc.remote[host][path] = &scaleRemoteFile{size: size, checksum: checksum}
	sc.mu.Unlock()
}

func (sc *scaleCluster) lookup(host, path string) *scaleRemoteFile {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.remote[host][path]
}

func (sc *scaleCluster) countEgress(host string, n int64) {
	sc.egressBytes.Add(n)
	sc.mu.Lock()
	sc.hostWireBytes[host] += n
	sc.mu.Unlock()
}

// ─── transport seam implementations (the mock) ──────────────────────

// upload is the TransferFunc for distribute: mirrors sshx.SFTPClient.Upload
// (mkdir -p remote dir, stream the local file) but lands the bytes on the
// virtual host and counts controller egress.
func (sc *scaleCluster) upload(ctx context.Context, src, endpoint string) error {
	host, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	if err := sc.sleep(ctx); err != nil {
		return err
	}

	// Stream the real local file: the controller genuinely reads and
	// hashes the bytes it pushes, like an SFTP upload would.
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("sim upload: open %s: %w", src, err)
	}
	defer f.Close()
	digest := sha256.New()
	size, err := io.Copy(digest, f)
	if err != nil {
		return fmt.Errorf("sim upload: read %s: %w", src, err)
	}

	if fail, corrupt := sc.injectFailure(host); fail {
		if corrupt {
			// Bytes arrived damaged; only the checksum verify hook can
			// detect this. Exercise exactly that path.
			sc.corruptions.Add(1)
			sc.countEgress(host, size)
			sc.store(host, remotePath, size, "corrupted:"+hex.EncodeToString(digest.Sum(nil)))
			return nil
		}
		sc.hardFails.Add(1)
		// Connection died mid-transfer after ~40% of the payload.
		sc.countEgress(host, size*2/5)
		return fmt.Errorf("sim upload %s: connection reset by peer (injected)", host)
	}

	sc.countEgress(host, size)
	sc.store(host, remotePath, size, hex.EncodeToString(digest.Sum(nil)))
	return nil
}

// verify is the VerifyFunc: compare the expected digest against what the
// virtual host actually stored - a real end-state check, not a rubber stamp.
func (sc *scaleCluster) verify(_ context.Context, endpoint, want string) error {
	host, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	rec := sc.lookup(host, remotePath)
	if rec == nil {
		return fmt.Errorf("sim verify %s: %s: no such file", host, remotePath)
	}
	if rec.checksum != want {
		return fmt.Errorf("sim verify %s: checksum mismatch: got %s want %s", host, rec.checksum, want)
	}
	return nil
}

// chmod is the ChmodFunc: records the mode on the virtual host.
func (sc *scaleCluster) chmod(_ context.Context, endpoint, mode string) error {
	host, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	sc.mu.Lock()
	if rec := sc.remote[host][remotePath]; rec != nil {
		rec.mode = mode
	}
	sc.mu.Unlock()
	return nil
}

// download is the CollectDownloadFunc for collect: mirrors
// sshx.SFTPClient.Download (mkdir -p local parent, stream to dst) with the
// same partial-write-on-error realism, and counts controller ingress.
func (sc *scaleCluster) download(ctx context.Context, endpoint, dst string) error {
	host, _, _, remotePath, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	if err := sc.sleep(ctx); err != nil {
		return err
	}
	if _, ok := sc.seedChecksum[host]; !ok {
		return fmt.Errorf("sim download %s: unknown host", host)
	}

	if fail, _ := sc.injectFailure(host); fail {
		sc.hardFails.Add(1)
		// Partial file lands locally, then the connection dies - a retry
		// must overwrite it cleanly.
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		f, ferr := os.Create(dst)
		if ferr != nil {
			return ferr
		}
		partial := &limitedWriter{max: sc.fileSize * 2 / 5}
		_, _ = sc.writeHostContent(host, io.MultiWriter(f, partial))
		cerr := f.Close()
		sc.ingressBytes.Add(partial.n)
		if cerr != nil {
			return cerr
		}
		return fmt.Errorf("sim download %s: connection reset by peer (injected)", host)
	}

	// Success: stream the virtual host's remote file to dst.
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("sim download: create %s: %w", dst, err)
	}
	digest := sha256.New()
	n, werr := sc.writeHostContent(host, io.MultiWriter(f, digest))
	cerr := f.Close()
	sc.ingressBytes.Add(n)
	if werr != nil {
		return werr
	}
	if cerr != nil {
		return cerr
	}
	if got := hex.EncodeToString(digest.Sum(nil)); got != sc.seedChecksum[host] {
		// Cannot happen unless the simulator itself is broken.
		return fmt.Errorf("sim download %s: generated content checksum drift for %s", host, remotePath)
	}
	return nil
}

// limitedWriter counts bytes up to max and then cuts the stream, like a
// connection dying mid-transfer (used for injected download failures).
type limitedWriter struct {
	max int64
	n   int64
}

var errWireCut = errors.New("sim wire cut")

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := w.max - w.n
	if remaining <= 0 {
		return 0, errWireCut
	}
	if int64(len(p)) > remaining {
		w.n = w.max
		return int(remaining), errWireCut
	}
	w.n += int64(len(p))
	return len(p), nil
}

// wireHostBytes returns the per-host controller-side byte totals (copy).
func (sc *scaleCluster) wireHostBytes() map[string]int64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	out := make(map[string]int64, len(sc.hostWireBytes))
	for k, v := range sc.hostWireBytes {
		out[k] = v
	}
	return out
}

// ─── assertions ─────────────────────────────────────────────────────

// requireSuccessRate asserts the >99.9% acceptance bar from CLAUDE.md
// Phase 4 using integer math (succ*1000 > total*999).
func requireSuccessRate(t *testing.T, label string, succ, total int) {
	t.Helper()
	if total == 0 {
		t.Fatalf("%s: total = 0", label)
	}
	if succ*1000 <= total*999 {
		t.Errorf("%s: success rate %.4f%% (%d/%d) violates the >99.9%% requirement",
			label, 100*float64(succ)/float64(total), succ, total)
	}
}

// ─── distribute at scale ────────────────────────────────────────────

func runScaleDistribute(t *testing.T, n int) {
	t.Helper()
	fileKB := scaleEnvInt(t, "OPS_SCALE_FILE_KB", 64)
	failRate := scaleEnvFloat(t, "OPS_SCALE_FAIL_RATE", 0.001)
	latency := time.Duration(scaleEnvInt(t, "OPS_SCALE_LATENCY_MS", 2)) * time.Millisecond

	// Real local source file on disk.
	dir := t.TempDir()
	src := filepath.Join(dir, "app-payload.bin")
	payload := make([]byte, fileKB*1024)
	rng := rand.New(rand.NewSource(7))
	rng.Read(payload)
	if err := os.WriteFile(src, payload, 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	wantChecksum, err := computeFileChecksum(src)
	if err != nil {
		t.Fatalf("source checksum: %v", err)
	}

	sc := newScaleCluster(t, n, fileKB, failRate, latency)

	// Install the simulator at the transport seam; the verify/chmod hooks
	// are package-level, so save/restore around the run.
	savedVerify, savedChmod := DefaultVerifyFunc, DefaultChmodFunc
	DefaultVerifyFunc = sc.verify
	DefaultChmodFunc = sc.chmod
	defer func() { DefaultVerifyFunc, DefaultChmodFunc = savedVerify, savedChmod }()

	targets := make([]DistributeTarget, 0, n)
	for _, host := range sc.hosts {
		targets = append(targets, DistributeTarget{
			Host: host,
			Port: 22,
			User: "root",
			Dest: "/opt/opslang/app-payload.bin",
		})
	}

	opts := DistributeOptions{
		Checksum: true,
		Mode:     "0644",
		Parallel: 64,
		Retries:  3,
		Timeout:  30 * time.Second,
	}

	result, err := DistributeWith(src, targets, opts, sc.upload)
	if err != nil {
		t.Fatalf("DistributeWith: %v", err)
	}

	// ── Aggregate shape: one result per host, no dupes, no losses. ──
	if result.Total != n {
		t.Errorf("Total = %d, want %d", result.Total, n)
	}
	if len(result.Results) != n {
		t.Errorf("len(Results) = %d, want %d", len(result.Results), n)
	}
	seen := make(map[string]int, n)
	for _, hr := range result.Results {
		seen[hr.Host]++
		if seen[hr.Host] > 1 {
			t.Errorf("duplicate result for host %s", hr.Host)
		}
	}
	for _, host := range sc.hosts {
		if seen[host] == 0 {
			t.Errorf("missing result for host %s", host)
		}
	}

	// ── Acceptance: success rate > 99.9% despite injected failures. ──
	requireSuccessRate(t, "distribute", result.Succeeded, result.Total)

	// ── Per-host correctness. ──
	for _, hr := range result.Results {
		if hr.Status != "success" {
			continue
		}
		if !hr.Changed {
			t.Errorf("host %s: Changed = false", hr.Host)
		}
		if hr.Size != int64(len(payload)) {
			t.Errorf("host %s: Size = %d, want %d", hr.Host, hr.Size, len(payload))
		}
		if hr.Checksum != wantChecksum {
			t.Errorf("host %s: Checksum = %q, want %q", hr.Host, hr.Checksum, wantChecksum)
		}
		// End-state on the virtual host: bytes + mode landed correctly.
		rec := sc.lookup(hr.Host, "/opt/opslang/app-payload.bin")
		if rec == nil {
			t.Errorf("host %s: no remote file stored", hr.Host)
			continue
		}
		if rec.checksum != wantChecksum {
			t.Errorf("host %s: remote checksum = %q, want %q (retry must repair corruption)", hr.Host, rec.checksum, wantChecksum)
		}
		if rec.size != int64(len(payload)) {
			t.Errorf("host %s: remote size = %d, want %d", hr.Host, rec.size, len(payload))
		}
		if rec.mode != "0644" {
			t.Errorf("host %s: remote mode = %q, want 0644", hr.Host, rec.mode)
		}
	}

	// ── Controller bandwidth: the current algorithm is direct fan-out
	// (no relay/dedup tier exists in the code), so the controller sends
	// one full copy per successful host plus bounded retry overhead.
	// Assert egress in [S*size, N*size*1.05]. ──
	size := int64(len(payload))
	egress := sc.egressBytes.Load()
	if min := int64(result.Succeeded) * size; egress < min {
		t.Errorf("controller egress %d < minimum %d (some host did not get its copy)", egress, min)
	}
	// Worst case: every host exhausts all retries -> N*retries*size. The
	// statistical bound with p=0.1% is ~N*1.001*size; allow 5% headroom
	// which still catches runaway retransmission loops.
	if max := int64(n) * size * 105 / 100; egress > max {
		t.Errorf("controller egress %d exceeds %d (5%% retry headroom over %d x %d bytes)", egress, max, n, size)
	}

	// ── Per-host wire bytes: each successful host saw >=1 and <=retries
	// full copies (no unbounded retransmission per host). ──
	perHost := sc.wireHostBytes()
	for _, hr := range result.Results {
		if hr.Status != "success" {
			continue
		}
		got := perHost[hr.Host]
		if got < size {
			t.Errorf("host %s: wire bytes %d < file size %d", hr.Host, got, size)
		}
		if got > int64(opts.Retries)*size {
			t.Errorf("host %s: wire bytes %d > retries*size %d", hr.Host, got, int64(opts.Retries)*size)
		}
	}

	// ── The failure injection must have actually fired so the retry
	// logic is genuinely exercised, not a pure happy path. ──
	if failRate > 0 && sc.hardFails.Load()+sc.corruptions.Load() == 0 {
		t.Errorf("failure injection never fired (hardFails=%d corruptions=%d); retry path untested",
			sc.hardFails.Load(), sc.corruptions.Load())
	}

	t.Logf("distribute scale: hosts=%d size=%dB failRate=%.4f latency=%v | succ=%d fail=%d hardFails=%d corruptions=%d egress=%dB (%.2fx payload) wall=%dms",
		n, size, failRate, latency, result.Succeeded, result.Failed,
		sc.hardFails.Load(), sc.corruptions.Load(), egress,
		float64(egress)/float64(size), result.DurationMs)
}

func TestScaleDistribute10000Hosts(t *testing.T) {
	if testing.Short() {
		t.Skip("scale simulation skipped in -short mode; run `make scale-test` for the full tier")
	}
	runScaleDistribute(t, scaleEnvInt(t, "OPS_SCALE_N", 10000))
}

func TestScaleDistributeCI(t *testing.T) {
	// Always-on CI gate tier.
	runScaleDistribute(t, scaleEnvInt(t, "OPS_SCALE_CI_N", 1000))
}

// ─── collect at scale ───────────────────────────────────────────────

func runScaleCollect(t *testing.T, n int) {
	t.Helper()
	fileKB := scaleEnvInt(t, "OPS_SCALE_FILE_KB", 64)
	failRate := scaleEnvFloat(t, "OPS_SCALE_FAIL_RATE", 0.001)
	latency := time.Duration(scaleEnvInt(t, "OPS_SCALE_LATENCY_MS", 2)) * time.Millisecond

	sc := newScaleCluster(t, n, fileKB, failRate, latency)

	destDir := t.TempDir()

	targets := make([]CollectTarget, 0, n)
	for _, host := range sc.hosts {
		targets = append(targets, CollectTarget{
			Host:   host,
			Port:   22,
			User:   "root",
			Source: "/var/log/app/service.log",
		})
	}

	result, err := CollectWith("/var/log/app/service.log", targets, CollectOptions{
		DestDir:  destDir,
		Parallel: 64,
		Retries:  3,
		Timeout:  30 * time.Second,
	}, sc.download)
	if err != nil {
		t.Fatalf("CollectWith: %v", err)
	}

	// ── Aggregate shape. ──
	if result.Total != n {
		t.Errorf("Total = %d, want %d", result.Total, n)
	}
	if len(result.Results) != n {
		t.Errorf("len(Results) = %d, want %d", len(result.Results), n)
	}
	seen := make(map[string]int, n)
	for _, hr := range result.Results {
		seen[hr.Host]++
		if seen[hr.Host] > 1 {
			t.Errorf("duplicate result for host %s", hr.Host)
		}
	}
	for _, host := range sc.hosts {
		if seen[host] == 0 {
			t.Errorf("missing result for host %s", host)
		}
	}

	// ── Acceptance: success rate > 99.9%. ──
	requireSuccessRate(t, "collect", result.Succeeded, result.Total)

	// ── Per-host archive: {destDir}/{host}/{basename} with the host's
	// own unique content (catches mis-attributed archives). ──
	for _, hr := range result.Results {
		if hr.Status != "success" {
			continue
		}
		wantDest := filepath.Join(destDir, hr.Host, "service.log")
		if hr.Dest != wantDest {
			t.Errorf("host %s: Dest = %q, want %q", hr.Host, hr.Dest, wantDest)
		}
		if hr.Size != sc.fileSize {
			t.Errorf("host %s: Size = %d, want %d", hr.Host, hr.Size, sc.fileSize)
		}
		wantChecksum := sc.seedChecksum[hr.Host]
		if hr.Checksum != wantChecksum {
			t.Errorf("host %s: Checksum = %q, want unique per-host digest %q", hr.Host, hr.Checksum, wantChecksum)
		}
		// The archived file on disk must match what the host really sent.
		info, err := os.Stat(wantDest)
		if err != nil {
			t.Errorf("host %s: archived file missing: %v", hr.Host, err)
			continue
		}
		if info.Size() != sc.fileSize {
			t.Errorf("host %s: archived size = %d, want %d", hr.Host, info.Size(), sc.fileSize)
		}
		got, err := computeFileChecksum(wantDest)
		if err != nil {
			t.Errorf("host %s: hash archived file: %v", hr.Host, err)
			continue
		}
		if got != wantChecksum {
			t.Errorf("host %s: archived checksum = %q, want %q (partial retry residue?)", hr.Host, got, wantChecksum)
		}
	}

	// ── Controller ingress: one full copy per successful host plus
	// bounded retry overhead ([S*size, N*size*1.05]). ──
	ingress := sc.ingressBytes.Load()
	if min := int64(result.Succeeded) * sc.fileSize; ingress < min {
		t.Errorf("controller ingress %d < minimum %d", ingress, min)
	}
	if max := int64(n) * sc.fileSize * 105 / 100; ingress > max {
		t.Errorf("controller ingress %d exceeds %d (5%% retry headroom)", ingress, max)
	}

	if failRate > 0 && sc.hardFails.Load() == 0 {
		t.Errorf("failure injection never fired (hardFails=%d); retry path untested", sc.hardFails.Load())
	}

	t.Logf("collect scale: hosts=%d size=%dB failRate=%.4f latency=%v | succ=%d fail=%d hardFails=%d ingress=%dB (%.2fx payload) wall=%dms",
		n, sc.fileSize, failRate, latency, result.Succeeded, result.Failed,
		sc.hardFails.Load(), ingress,
		float64(ingress)/float64(sc.fileSize), result.DurationMs)
}

func TestScaleCollect10000Hosts(t *testing.T) {
	if testing.Short() {
		t.Skip("scale simulation skipped in -short mode; run `make scale-test` for the full tier")
	}
	runScaleCollect(t, scaleEnvInt(t, "OPS_SCALE_N", 10000))
}

func TestScaleCollectCI(t *testing.T) {
	// Always-on CI gate tier.
	runScaleCollect(t, scaleEnvInt(t, "OPS_SCALE_CI_N", 1000))
}
