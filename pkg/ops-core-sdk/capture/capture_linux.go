//go:build linux

// Package capture implements passive packet capture over a Linux AF_PACKET
// raw socket (pure Go: no libpcap, no cgo, keeps CGO_ENABLED=0 cross builds
// working). It exists so opsctl/ops-runner can do tcpdump-style triage —
// retransmissions, RSTs, zero windows, ARP — and emit STRUCTURED results,
// without shelling out to external tools.
//
// Privileges: opening an ETH_P_ALL raw socket requires CAP_NET_RAW. Run the
// binary under sudo or grant the binary the capability:
//
//	sudo setcap cap_net_raw+ep /path/to/opsctl
package capture

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Options controls one capture window.
type Options struct {
	// Iface to listen on ("eth0"). Empty binds to all interfaces (any).
	Iface string
	// Seconds bounds the capture window; 0 means 5.
	Seconds int
	// MaxPkts stops the capture after this many frames; 0 means 200.
	MaxPkts int
	// SnapLen caps stored bytes per frame for the pcap dump; headers need
	// <= 74 bytes, so 256 is plenty for diagnosis. Clamped to [64, 262144].
	SnapLen int
	// Port filters post-parse on either TCP/UDP port when > 0.
	Port int
	// Host filters post-parse when non-empty (dotted IPv4, either side).
	Host string
	// PcapPath writes a classic libpcap file (Wireshark/tcpdump readable)
	// from the retained raw frames. "" skips writing.
	PcapPath string
}

// Packet is one parsed frame in structured form.
type Packet struct {
	No        int     `json:"no"`
	TSUsec    int64   `json:"ts_usec"` // absolute unix microseconds
	MsFrom0   float64 `json:"ms_from_start"`
	SrcMac    string  `json:"src_mac"`
	DstMac    string  `json:"dst_mac"`
	Ethertype string  `json:"ethertype"` // IPv4 | IPv6 | ARP | VLAN | Other
	Proto     string  `json:"proto"`     // TCP|UDP|ICMP|ARP|IPv6|Other
	SrcIP     string  `json:"src_ip,omitempty"`
	DstIP     string  `json:"dst_ip,omitempty"`
	SrcPort   int     `json:"src_port,omitempty"`
	DstPort   int     `json:"dst_port,omitempty"`
	Flags     string  `json:"tcp_flags,omitempty"` // tcpdump style: [S.] [P.] [R]
	Seq       uint32  `json:"seq,omitempty"`
	Ack       uint32  `json:"ack,omitempty"`
	Win       uint16  `json:"win,omitempty"`
	PayLen    int     `json:"payload_len,omitempty"`
	WireLen   int     `json:"wire_len"`
}

// Result is the structured outcome of a capture window.
type Result struct {
	Iface         string         `json:"iface"`
	StartedAt     int64          `json:"started_at_unix"`
	DurationMs    int64          `json:"duration_ms"`
	Captured      int            `json:"captured"`     // frames read
	Matched       int            `json:"matched"`      // frames passing host/port filter
	KernelDrops   uint32         `json:"kernel_drops"` // NIC->kernel ring drops during window
	ProtoCounts   map[string]int `json:"proto_counts"`
	TCPEvents     TCPEventCounts `json:"tcp_events"`
	Conversations []string       `json:"conversations"` // top flows by packet count
	PcapPath      string         `json:"pcap_path"`
	// PCapB64/PCapLocalPath appear only for "local:" targets under remote
	// execution: the payload rides inside the result and the controller
	// writes it to the operator-side path (see embed.go).
	PCapB64       string   `json:"__pcap_b64,omitempty"`
	PCapLocalPath string   `json:"__pcap_local_path,omitempty"`
	Packets       []Packet `json:"packets"`
}

// TCPEventCounts distills exactly the anomalies the runbook looks for.
type TCPEventCounts struct {
	Syn        int `json:"syn"`         // pure SYN (connection attempts)
	SynAck     int `json:"syn_ack"`     // SYN+ACK (accepted attempts)
	Rst        int `json:"rst"`         // resets (refusals/idle-timeouts)
	Fin        int `json:"fin"`         // graceful closes
	ZeroWindow int `json:"zero_window"` // receiver buffer full
}

// Capture runs one passive capture window.
func Capture(opts Options) (*Result, error) {
	if opts.Seconds <= 0 {
		opts.Seconds = 5
	}
	if opts.MaxPkts <= 0 {
		opts.MaxPkts = 200
	}
	if opts.SnapLen == 0 {
		opts.SnapLen = 256
	}
	if opts.SnapLen < 64 {
		opts.SnapLen = 64
	}
	if opts.SnapLen > 262144 {
		opts.SnapLen = 262144
	}
	if opts.MaxPkts > 8000 {
		opts.MaxPkts = 8000 // memory guard: 8000 * snaplen <= ~2MB at default
	}

	var ifi *net.Interface
	if opts.Iface == "" || opts.Iface == "any" {
		// bind to "any": ifindex 0 with protocol ETH_P_ALL (SLL pseudo-iface)
		ifi = &net.Interface{Index: 0}
	} else {
		var err error
		ifi, err = net.InterfaceByName(opts.Iface)
		if err != nil {
			return nil, fmt.Errorf("net.capture: interface %q: %w", opts.Iface, err)
		}
	}

	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, int(htons(unix.ETH_P_ALL)))
	if err != nil {
		return nil, fmt.Errorf("net.capture: raw socket (need CAP_NET_RAW/root): %w", err)
	}
	defer unix.Close(fd)

	sa := &unix.SockaddrLinklayer{Ifindex: ifi.Index, Protocol: htons(unix.ETH_P_ALL)}
	if err := unix.Bind(fd, sa); err != nil {
		return nil, fmt.Errorf("net.capture: bind %s: %w", opts.Iface, err)
	}

	// 250ms receive timeout turns blocking reads into ticks we can count.
	tv := unix.NsecToTimeval(250 * time.Millisecond.Nanoseconds())
	if err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv); err != nil {
		return nil, fmt.Errorf("net.capture: rcvtimeo: %w", err)
	}

	start := time.Now()
	deadline := start.Add(time.Duration(opts.Seconds) * time.Second)

	res := &Result{
		Iface:       opts.Iface,
		StartedAt:   start.Unix(),
		ProtoCounts: map[string]int{},
	}
	var frames []capturedFrame

	buf := make([]byte, opts.SnapLen)
	for time.Now().Before(deadline) && len(frames) < opts.MaxPkts {
		n, _, err := unix.Recvfrom(fd, buf, 0)
		now := time.Now()
		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				continue // read tick timeout
			}
			if err == unix.EINTR {
				continue // signal interrupted the read: keep capturing
			}
			return nil, fmt.Errorf("net.capture: recvfrom: %w", err)
		}
		if n <= 0 {
			continue
		}
		cf := capturedFrame{
			raw:     append([]byte(nil), buf[:n]...),
			sec:     now.Unix(),
			usec:    int64(now.Nanosecond() / 1000),
			origLen: n,
		}
		frames = append(frames, cf)
	}
	dur := time.Since(start)
	res.DurationMs = dur.Milliseconds()

	// Kernel ring statistics: how many frames the NIC/driver dropped before
	// user space ever saw them. Honest signal that results are a sample.
	if v, errno := unix.GetsockoptTpacketStats(fd, unix.SOL_PACKET, unix.PACKET_STATISTICS); errno == nil && v != nil {
		res.KernelDrops = v.Drops
	}

	for _, cf := range frames {
		pkt := ParseFrame(cf.raw)
		if pkt == nil {
			continue
		}
		pkt.No = len(res.Packets) + 1
		pkt.TSUsec = cf.sec*1e6 + cf.usec
		pkt.MsFrom0 = float64(cf.sec-start.Unix())*1000 + float64(cf.usec)/1000
		pkt.WireLen = cf.origLen

		res.Captured++

		if !passFilter(pkt, opts.Host, opts.Port) {
			continue
		}
		res.Matched++
		res.ProtoCounts[pkt.Proto]++
		collectTCPEvents(res, pkt)
		maybeNoteConversation(res, pkt)
		res.Packets = append(res.Packets, *pkt)
	}

	topConversations(res)

	if opts.PcapPath != "" {
		if werr := WritePcap(opts.PcapPath, frames, opts.SnapLen); werr != nil {
			return res, fmt.Errorf("net.capture succeeded but pcap write failed: %w", werr)
		}
		res.PcapPath = opts.PcapPath
	}
	return res, nil
}

// Run is the positional-argument shape the AOT code generator emits calls
// for. It wraps the Options-based Capture and owns "local:" handling for
// the AOT context: a compiled binary executes on the CAPTURING host, so a
// local: target stages here, embeds into the result, and cleans up - the
// controller materializes it afterwards.
func Run(iface string, seconds int, maxPackets int, pcapPath string) (Result, error) {
	local, userTarget := SplitPcapTarget(pcapPath)
	// On the capturing host we must never try to materialize the caller's
	// workstation path; stage to a temp file, embed the bytes, and let the
	// controller write userTarget. The temp file is removed afterwards.
	fsPath := userTarget
	if local {
		f, err := os.CreateTemp("", "ops-cap-*.pcap")
		if err != nil {
			return Result{}, fmt.Errorf("net.capture local temp: %w", err)
		}
		fsPath = f.Name()
		f.Close()
	}
	r, err := Capture(Options{
		Iface:    iface,
		Seconds:  seconds,
		MaxPkts:  maxPackets,
		PcapPath: fsPath,
	})
	if err != nil {
		if local {
			os.Remove(fsPath)
		}
		return Result{}, err
	}
	if local {
		if lerr := MaterializeLocal(r, fsPath, userTarget); lerr != nil {
			return Result{}, lerr
		}
	}
	return *r, nil
}
