package causal

import (
	"os"
	"path/filepath"
	"testing"

	opsnet "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/net"
)

func TestParseProcStat(t *testing.T) {
	stat := "123 (java worker) S 42"
	for i := 0; i < 17; i++ {
		stat += " 0"
	}
	stat += " 777"
	pid, ppid, start, ok := parseProcStat(stat)
	if !ok || pid != 123 || ppid != 42 || start != 777 {
		t.Fatalf("parseProcStat() = %d, %d, %d, %v", pid, ppid, start, ok)
	}
}

func TestTracePIDFromRootBuildsParentChain(t *testing.T) {
	root := t.TempDir()
	writeProcFixture(t, root, 100, 1, "systemd", "/sbin/init", "systemd --system")
	writeProcFixture(t, root, 200, 100, "sshd", "/usr/sbin/sshd", "sshd: user")
	writeProcFixture(t, root, 300, 200, "java", "/usr/bin/java", "/usr/bin/java -jar app.jar")

	trace, err := tracePIDFromRoot(root, 300)
	if err != nil {
		t.Fatalf("tracePIDFromRoot() error: %v", err)
	}
	if len(trace.Nodes) != 3 || len(trace.Edges) != 2 {
		t.Fatalf("unexpected chain: nodes=%d edges=%d", len(trace.Nodes), len(trace.Edges))
	}
	if trace.Nodes[0].PID != 300 || trace.Nodes[1].PID != 200 || trace.Nodes[2].PID != 100 {
		t.Fatalf("unexpected node order: %+v", trace.Nodes)
	}
	if trace.Edges[0].Relation != "parent" || trace.Edges[0].FromPID != 300 || trace.Edges[0].ToPID != 200 {
		t.Fatalf("unexpected edge: %+v", trace.Edges[0])
	}
}

func TestTraceFileFromRootFindsOpenDescriptor(t *testing.T) {
	root := t.TempDir()
	writeProcFixture(t, root, 100, 1, "systemd", "/sbin/init", "systemd --system")
	writeProcFixture(t, root, 300, 100, "java", "/usr/bin/java", "/usr/bin/java -jar app.jar")
	if err := os.MkdirAll(filepath.Join(root, "300", "fd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/opt/app/app.jar", filepath.Join(root, "300", "fd", "7")); err != nil {
		t.Fatal(err)
	}
	traces, err := traceFileFromRoot(root, "/opt/app/app.jar")
	if err != nil {
		t.Fatalf("traceFileFromRoot() error: %v", err)
	}
	if len(traces) != 1 || traces[0].PID != 300 || traces[0].FD != 7 || len(traces[0].Trace.Nodes) != 2 {
		t.Fatalf("unexpected file traces: %+v", traces)
	}
}

func TestTraceContainerFromRootMatchesCgroup(t *testing.T) {
	root := t.TempDir()
	writeProcFixture(t, root, 100, 1, "systemd", "/sbin/init", "systemd --system")
	writeProcFixture(t, root, 300, 100, "java", "/usr/bin/java", "/usr/bin/java -jar app.jar")
	if err := os.WriteFile(filepath.Join(root, "300", "cgroup"), []byte("0::/docker/abcdef123456\n"), 0644); err != nil {
		t.Fatal(err)
	}
	traces, err := traceContainerFromRoot(root, "abcdef123456")
	if err != nil {
		t.Fatalf("traceContainerFromRoot() error: %v", err)
	}
	if len(traces) != 1 || traces[0].Nodes[0].ContainerID != "abcdef123456" {
		t.Fatalf("unexpected container traces: %+v", traces)
	}
}

func TestTracePortFromConnectionsDeduplicatesPIDs(t *testing.T) {
	connections := []opsnet.ConnectionInfo{
		{LocalAddr: "0.0.0.0:8080", Pid: 300, ProcessName: "java", Protocol: "tcp"},
		{LocalAddr: "[::]:8080", Pid: 300, ProcessName: "java", Protocol: "tcp6"},
		{LocalAddr: "127.0.0.1:9090", Pid: 400, ProcessName: "other", Protocol: "tcp"},
	}
	matched := matchingPortConnections(connections, 8080)
	if len(matched) != 2 || matched[0].PID != 300 || matched[1].PID != 300 {
		t.Fatalf("unexpected port matches: %+v", matched)
	}
}

func TestTraceContainerRejectsMalformedID(t *testing.T) {
	if _, err := TraceContainer("not-hex"); err == nil {
		t.Fatal("TraceContainer() must reject malformed IDs")
	}
}

func writeProcFixture(t *testing.T, root string, pid, ppid int32, name, exe, cmdline string) {
	t.Helper()
	dir := filepath.Join(root, itoa(pid))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	stat := itoa(pid) + " (" + name + ") S " + itoa(ppid)
	for i := 0; i < 17; i++ {
		stat += " 0"
	}
	stat += " 123"
	if err := os.WriteFile(filepath.Join(dir, "stat"), []byte(stat), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "comm"), []byte(name+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(cmdline), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "status"), []byte("Uid:\t1000\t1000\t1000\t1000\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(exe, filepath.Join(dir, "exe")); err != nil {
		t.Fatal(err)
	}
}

func itoa(v int32) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	b := make([]byte, 0, 12)
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
