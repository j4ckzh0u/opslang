package capture

// Shared types for the capture package. Kept in a platform-neutral file so
// the Linux (AF_PACKET) and all-other-platforms (stub) builds cannot drift:
// Packet / Result / Options / TCPEventCounts are identical everywhere so
// generated code and JSON consumers stay compatible.

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

	// liveConversations accumulates per-flow counts during one Capture; it
	// is internal scratch space (not serialized) and makes Capture
	// reentrant/concurrency-safe instead of using a hidden global.
	liveConversations map[string]int
}

// TCPEventCounts distills exactly the anomalies the runbook looks for.
type TCPEventCounts struct {
	Syn        int `json:"syn"`         // pure SYN (connection attempts)
	SynAck     int `json:"syn_ack"`     // SYN+ACK (accepted attempts)
	Rst        int `json:"rst"`         // resets (refusals/idle-timeouts)
	Fin        int `json:"fin"`         // graceful closes
	ZeroWindow int `json:"zero_window"` // receiver buffer full
}
