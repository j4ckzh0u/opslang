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
			os.Remove(fsPath)
			return Result{}, lerr
		}
	}
	return *r, nil
}
