package capture

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// synth builders ---------------------------------------------------------------

func eth(srcMAC, dstMAC []byte, ethertype uint16, payload []byte) []byte {
	out := make([]byte, 14+len(payload))
	copy(out[0:6], dstMAC)
	copy(out[6:12], srcMAC)
	binary.BigEndian.PutUint16(out[12:14], ethertype)
	copy(out[14:], payload)
	return out
}

func ipv4(proto byte, src, dst [4]byte, l4 []byte) []byte {
	b := make([]byte, 20+len(l4))
	b[0] = 0x45
	total := len(b)
	binary.BigEndian.PutUint16(b[2:4], uint16(total))
	b[8] = 64
	b[9] = proto
	copy(b[12:16], src[:])
	copy(b[16:20], dst[:])
	copy(b[20:], l4)
	return b
}

func tcpHdr(sport, dport int, seq, ack uint32, flags byte, win uint16, payload []byte) []byte {
	b := make([]byte, 20+len(payload))
	binary.BigEndian.PutUint16(b[0:2], uint16(sport))
	binary.BigEndian.PutUint16(b[2:4], uint16(dport))
	binary.BigEndian.PutUint32(b[4:8], seq)
	binary.BigEndian.PutUint32(b[8:12], ack)
	b[12] = 5 << 4
	b[13] = flags
	binary.BigEndian.PutUint16(b[14:16], win)
	copy(b[20:], payload)
	return b
}

// tests ------------------------------------------------------------------------

var (
	srcA = [4]byte{10, 0, 0, 1}
	dstB = [4]byte{10, 0, 0, 2}
	mac1 = []byte{2, 0, 0, 0, 0, 1}
	mac2 = []byte{2, 0, 0, 0, 0, 2}
)

func TestParseFrameTCPSyn(t *testing.T) {
	frame := eth(mac1, mac2, 0x0800, ipv4(6, srcA, dstB, tcpHdr(45678, 3306, 123, 0, 0x02, 29200, nil)))
	p := ParseFrame(frame)
	if p == nil {
		t.Fatal("nil packet")
	}
	if p.Proto != "TCP" || p.Ethertype != "IPv4" {
		t.Fatalf("proto/et = %s/%s", p.Proto, p.Ethertype)
	}
	if p.SrcIP != "10.0.0.1" || p.DstIP != "10.0.0.2" {
		t.Fatalf("ips %s->%s", p.SrcIP, p.DstIP)
	}
	if p.SrcPort != 45678 || p.DstPort != 3306 {
		t.Fatalf("ports %d->%d", p.SrcPort, p.DstPort)
	}
	if p.Flags != "[S]" {
		t.Fatalf("flags %q", p.Flags)
	}
	if p.Seq != 123 || p.Win != 29200 {
		t.Fatalf("seq/win %d/%d", p.Seq, p.Win)
	}
}

func TestParseFrameFlagShapes(t *testing.T) {
	cases := map[byte]string{
		0x12: "[S.]", // SYN+ACK
		0x18: "[P.]", // PSH+ACK
		0x11: "[F.]", // FIN+ACK
		0x04: "[R]",  // bare RST
		0x10: "[.]",  // pure ACK
	}
	for fl, want := range cases {
		frame := eth(mac1, mac2, 0x0800, ipv4(6, srcA, dstB, tcpHdr(1, 2, 1, 1, fl, 100, nil)))
		p := ParseFrame(frame)
		if p.Flags != want {
			t.Errorf("flags %#x = %q want %q", fl, p.Flags, want)
		}
	}
}

func TestParseFrameUDPAndICMP(t *testing.T) {
	udp := make([]byte, 13) // 8 hdr + 5 payload
	binary.BigEndian.PutUint16(udp[0:2], 53)
	binary.BigEndian.PutUint16(udp[2:4], 40000)
	binary.BigEndian.PutUint16(udp[4:6], 13) // UDP length includes header
	frame := eth(mac1, mac2, 0x0800, ipv4(17, srcA, dstB, udp))
	p := ParseFrame(frame)
	if p.Proto != "UDP" || p.SrcPort != 53 || p.PayLen != 5 {
		t.Fatalf("udp parse: %+v", p)
	}

	icmp := make([]byte, 9) // echo id/seq + 1 data byte (hdr=8)
	icmp[0] = 8
	frame = eth(mac1, mac2, 0x0800, ipv4(1, srcA, dstB, icmp))
	p = ParseFrame(frame)
	if p.Proto != "ICMP" || p.PayLen != 1 {
		t.Fatalf("icmp parse: %+v", p)
	}
}

func TestParseFrameARP(t *testing.T) {
	ar := make([]byte, 28)
	binary.BigEndian.PutUint16(ar[2:4], 0x0800) // protocol type = IPv4
	binary.BigEndian.PutUint16(ar[6:8], 1)      // operation = request
	copy(ar[14:18], srcA[:])
	copy(ar[24:28], dstB[:])
	frame := eth(mac1, mac2, 0x0806, ar)
	p := ParseFrame(frame)
	if p.Ethertype != "ARP" || p.Flags != "[request]" || p.SrcIP != "10.0.0.1" {
		t.Fatalf("arp parse: %+v", p)
	}
}

func TestParseFrameEdgeCases(t *testing.T) {
	if ParseFrame(nil) != nil {
		t.Error("nil frame must give nil")
	}
	if ParseFrame([]byte{1, 2, 3}) != nil {
		t.Error("runt frame must give nil")
	}
	// IPv4 claims proto TCP but only 10 bytes of L4 -> TCP-trunc, nil
	trunc := eth(mac1, mac2, 0x0800, ipv4(6, srcA, dstB, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))
	if p := ParseFrame(trunc); p != nil && p.Proto == "TCP" {
		t.Errorf("truncated TCP should not parse as clean TCP: %+v", p)
	}
	// unknown ethertype
	unk := eth(mac1, mac2, 0x9999, []byte{1, 2})
	if p := ParseFrame(unk); p == nil || p.Ethertype != "Other" {
		t.Errorf("unknown et: %+v", p)
	}
}

func TestPassFilter(t *testing.T) {
	frame := eth(mac1, mac2, 0x0800, ipv4(6, srcA, dstB, tcpHdr(1234, 80, 0, 0, 0x10, 100, nil)))
	p := ParseFrame(frame)
	if !passFilter(p, "", 0) {
		t.Error("empty filter passes all")
	}
	if !passFilter(p, "10.0.0.2", 0) {
		t.Error("dst side match")
	}
	if passFilter(p, "10.9.9.9", 0) {
		t.Error("unrelated host must be filtered out")
	}
	if !passFilter(p, "", 80) {
		t.Error("dst port match")
	}
	if passFilter(p, "", 443) {
		t.Error("wrong port filtered")
	}
}

func TestWritePcapRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.pcap")
	frames := []capturedFrame{}
	f1 := eth(mac1, mac2, 0x0800, ipv4(6, srcA, dstB, tcpHdr(1, 2, 1, 1, 0x02, 99, nil)))
	frames = append(frames, capturedFrame{raw: f1, sec: 1724736000, usec: 500, origLen: len(f1)})
	f2 := eth(mac2, mac1, 0x0806, make([]byte, 28))
	frames = append(frames, capturedFrame{raw: f2, sec: 1724736001, usec: 42, origLen: len(f2)})

	if err := WritePcap(path, frames, 256); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0xd4 || got[1] != 0xc3 || got[2] != 0xb2 || got[3] != 0xa1 {
		t.Fatalf("bad magic % x", got[0:4])
	}
	// global hdr 24B; per rec 16B header; verify count via walking records
	n := 0
	off := 24
	for off+16 <= len(got) {
		incl := int(got[off+8]) | int(got[off+9])<<8 | int(got[off+10])<<16 | int(got[off+11])<<24
		orig := int(got[off+12]) | int(got[off+13])<<8 | int(got[off+14])<<16 | int(got[off+15])<<24
		if incl != orig {
			t.Errorf("incl!=orig at rec %d: %d vs %d", n, incl, orig)
		}
		off += 16 + incl
		n++
	}
	if n != 2 {
		t.Fatalf("records=%d want 2", n)
	}
	if !bytes.HasPrefix(got[24+16:], f1) {
		t.Error("first frame bytes mangled")
	}
}

func TestCaptureInvalidIface(t *testing.T) {
	if _, err := Capture(Options{Iface: "does-not-exist-xyz"}); err == nil {
		t.Fatal("bad iface must error")
	}
}

func TestMaybeNoteConversationAndTop(t *testing.T) {
	r := &Result{}
	mk := func(sip string, sport int, dip string, dport int) *Packet {
		return &Packet{Proto: "TCP", SrcIP: sip, SrcPort: sport, DstIP: dip, DstPort: dport}
	}
	// flow A (client 10.0.0.1:4000 -> server 10.0.0.9:443) 3 packets.
	// 4000 < 443 is false -> else branch -> key "10.0.0.9:443<10.0.0.1:4000".
	for i := 0; i < 3; i++ {
		maybeNoteConversation(r, mk("10.0.0.1", 4000, "10.0.0.9", 443))
	}
	// reverse-direction packet with swapped ports is a DISTINCT key under this
	// (non-canonicalizing) keying scheme.
	maybeNoteConversation(r, mk("10.0.0.9", 443, "10.0.0.1", 4000))
	// a UDP flow, 1 packet (53 < 5353 -> A>B branch)
	maybeNoteConversation(r, &Packet{Proto: "UDP", SrcIP: "1.1.1.1", SrcPort: 53, DstIP: "2.2.2.2", DstPort: 5353})
	// non-TCP/UDP or port 0 must not count
	maybeNoteConversation(r, &Packet{Proto: "ICMP", SrcIP: "7.7.7.7", SrcPort: 0})
	maybeNoteConversation(r, &Packet{Proto: "TCP", SrcIP: "3.3.3.3", SrcPort: 0, DstIP: "4.4.4.4", DstPort: 80})

	topConversations(r)
	want := []string{
		"10.0.0.9:443<10.0.0.1:4000 TCP x3",
		"1.1.1.1:53>2.2.2.2:5353 UDP x1",
		"10.0.0.9:443>10.0.0.1:4000 TCP x1",
	}
	if len(r.Conversations) != len(want) {
		t.Fatalf("want %d conversations, got %v", len(want), r.Conversations)
	}
	for i := range want {
		if r.Conversations[i] != want[i] {
			t.Fatalf("conv[%d]=%q want %q", i, r.Conversations[i], want[i])
		}
	}
	// liveConversations cleared (no cross-call bleed)
	if r.liveConversations != nil {
		t.Fatal("liveConversations must reset after topConversations")
	}
}

func TestCollectTCPEventsFlags(t *testing.T) {
	// [S.] -> SynAck (not pure Syn)
	r := &Result{}
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[S.]", Win: 65000})
	if r.TCPEvents.SynAck != 1 || r.TCPEvents.Syn != 0 {
		t.Fatalf("synack miscount: %+v", r.TCPEvents)
	}
	// [S] pure -> Syn
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[S]", Win: 65000})
	if r.TCPEvents.Syn != 1 {
		t.Fatalf("syn miscount: %+v", r.TCPEvents)
	}
	// [R] + [F]
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[R]"})
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[F.]"})
	if r.TCPEvents.Rst != 1 || r.TCPEvents.Fin != 1 {
		t.Fatalf("rst/fin miscount: %+v", r.TCPEvents)
	}
	// zero-window: win=0 with ACK, no S/R/F
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[.]", Win: 0})
	if r.TCPEvents.ZeroWindow != 1 {
		t.Fatalf("zero-window miscount: %+v", r.TCPEvents)
	}
	// SYN with win 0 must NOT count as zero-window
	collectTCPEvents(r, &Packet{Proto: "TCP", Flags: "[S]", Win: 0})
	if r.TCPEvents.ZeroWindow != 1 {
		t.Fatalf("SYN must not be zero-window: %+v", r.TCPEvents)
	}
}

func TestConversationsConcurrencySafe(t *testing.T) {
	// After the refactor, conversation state lives on each Result, so many
	// goroutines counting into separate Results must not interfere. This
	// would race (or fail under -race) on the old package-global map.
	const n = 200
	var wg sync.WaitGroup
	ok := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := &Result{}
			maybeNoteConversation(r, &Packet{Proto: "TCP", SrcIP: "10.0.0.1", SrcPort: 1000 + i, DstIP: "10.0.0.9", DstPort: 443})
			topConversations(r)
			ok <- r.Conversations[0]
		}(i)
	}
	wg.Wait()
	close(ok)
	count := 0
	for range ok {
		count++
	}
	if count != n {
		t.Fatalf("lost results: %d != %d", count, n)
	}
}
