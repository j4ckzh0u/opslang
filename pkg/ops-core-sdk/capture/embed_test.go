package capture

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitPcapTarget(t *testing.T) {
	local, p := SplitPcapTarget("local:/tmp/a.pcap")
	if !local || p != "/tmp/a.pcap" {
		t.Fatalf("local split: %v %q", local, p)
	}
	local, p = SplitPcapTarget("/tmp/b.pcap")
	if local || p != "/tmp/b.pcap" {
		t.Fatalf("plain split: %v %q", local, p)
	}
}

func TestMaterializeLocalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	staged := filepath.Join(dir, "staged.pcap")
	payload := []byte("PCAPPAYLOAD123")
	if err := os.WriteFile(staged, payload, 0644); err != nil {
		t.Fatal(err)
	}
	res := &Result{PcapPath: staged}
	final := filepath.Join(dir, "final.pcap")
	if err := MaterializeLocal(res, staged, final); err != nil {
		t.Fatal(err)
	}
	if res.PCapB64 == "" || res.PCapLocalPath != final || res.PcapPath != "" {
		t.Fatalf("result fields wrong: %+v", res)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatal("staged file must be removed after embedding")
	}

	// controller side
	tree := map[string]interface{}{
		"cap": map[string]interface{}{
			KeyPCapB64:       res.PCapB64,
			KeyPCapLocalPath: final,
		},
	}
	saved, warns := SaveEmbedded(tree)
	if len(warns) != 0 || len(saved) != 1 {
		t.Fatalf("save: saved=%v warns=%v", saved, warns)
	}
	got, err := os.ReadFile(saved[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch: %q", got)
	}
	m := tree["cap"].(map[string]interface{})
	if _, still := m[KeyPCapB64]; still {
		t.Fatal("__pcap_b64 must be stripped from result")
	}
}

func TestSaveEmbeddedOversizeAndGarbage(t *testing.T) {
	big := make([]byte, MaxEmbedBytes+1)
	res := Result{}
	if err := MaterializeLocal(&res, writeTempBig(t, big), "/x.pcap"); err == nil {
		t.Fatal("oversize must refuse")
	}
	tree := map[string]interface{}{
		KeyPCapB64:       "!!!not-base64!!!",
		KeyPCapLocalPath: "/tmp/should-never-be-written.pcap",
	}
	saved, warns := SaveEmbedded(map[string]interface{}{"x": tree})
	if len(saved) != 0 || len(warns) != 1 {
		t.Fatalf("garbage b64: saved=%v warns=%v", saved, warns)
	}
	if _, err := os.Stat("/tmp/should-never-be-written.pcap"); !os.IsNotExist(err) {
		t.Fatal("must not write on decode failure")
	}
}

func TestSaveEmbeddedNestedList(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "n.pcap")
	raw := []byte{1, 2, 3}
	tree := map[string]interface{}{
		"hosts": []interface{}{
			map[string]interface{}{KeyPCapB64: b64(raw), KeyPCapLocalPath: target},
		},
	}
	saved, _ := SaveEmbedded(tree)
	if len(saved) != 1 {
		t.Fatalf("nested list miss: %v", saved)
	}
	got, _ := os.ReadFile(target)
	if len(got) != 3 {
		t.Fatalf("bytes: %v", got)
	}
}

// helpers

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func writeTempBig(t *testing.T, content []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "big.pcap")
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	return p
}
