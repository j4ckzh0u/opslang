//go:build !linux

// Package capture: passive packet capture. Linux-only by design (AF_PACKET);
// on every other platform this stub returns an explicit error so scripts
// fail loudly instead of silently, and CGO stays disabled everywhere.
package capture

import "errors"

// Options mirrors the Linux build; accepted but unusable here.
type Options struct {
	Iface    string
	Seconds  int
	MaxPkts  int
	SnapLen  int
	Port     int
	Host     string
	PcapPath string
}

// Packet / Result mirror the Linux field shapes so generated code and JSON
// consumers stay identical across platforms.
type Packet struct {
	No        int     `json:"no"`
	TSUsec    int64   `json:"ts_usec"`
	MsFrom0   float64 `json:"ms_from_start"`
	SrcMac    string  `json:"src_mac"`
	DstMac    string  `json:"dst_mac"`
	Ethertype string  `json:"ethertype"`
	Proto     string  `json:"proto"`
	SrcIP     string  `json:"src_ip,omitempty"`
	DstIP     string  `json:"dst_ip,omitempty"`
	SrcPort   int     `json:"src_port,omitempty"`
	DstPort   int     `json:"dst_port,omitempty"`
	Flags     string  `json:"tcp_flags,omitempty"`
	Seq       uint32  `json:"seq,omitempty"`
	Ack       uint32  `json:"ack,omitempty"`
	Win       uint16  `json:"win,omitempty"`
	PayLen    int     `json:"payload_len,omitempty"`
	WireLen   int     `json:"wire_len"`
}

type TCPEventCounts struct {
	Syn        int `json:"syn"`
	SynAck     int `json:"syn_ack"`
	Rst        int `json:"rst"`
	Fin        int `json:"fin"`
	ZeroWindow int `json:"zero_window"`
}

type Result struct {
	Iface         string         `json:"iface"`
	StartedAt     int64          `json:"started_at_unix"`
	DurationMs    int64          `json:"duration_ms"`
	Captured      int            `json:"captured"`
	Matched       int            `json:"matched"`
	KernelDrops   uint32         `json:"kernel_drops,omitempty"`
	ProtoCounts   map[string]int `json:"proto_counts"`
	TCPEvents     TCPEventCounts `json:"tcp_events"`
	Conversations []string       `json:"conversations,omitempty"`
	PcapPath      string         `json:"pcap_path,omitempty"`
	Packets       []Packet       `json:"packets"`
}

var errUnsupported = errors.New("net.capture: passive capture requires Linux (AF_PACKET); rebuild the tool on a Linux host")

func Capture(Options) (*Result, error) { return nil, errUnsupported }

// Run mirrors the Linux signature so generated code compiles everywhere;
// execution returns errUnsupported.
func Run(string, int, int, string) (Result, error) {
	return Result{}, errUnsupported
}
