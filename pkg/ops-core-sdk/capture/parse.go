package capture

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
)

// capturedFrame keeps the retained raw bytes plus their timestamps so a
// pcap file can be written without re-reading anything.
type capturedFrame struct {
	raw     []byte
	sec     int64
	usec    int64
	origLen int
}

// ParseFrame decodes an Ethernet frame into structured fields. Returns nil
// for anything it cannot at least place at ether/ARP/IPv4 granularity
// (runts, non-IP non-ARP payloads). Never panics on short input.
func ParseFrame(b []byte) *Packet {
	if len(b) < 14 {
		return nil
	}
	p := &Packet{
		DstMac: macString(b[0:6]),
		SrcMac: macString(b[6:12]),
	}
	et := binary.BigEndian.Uint16(b[12:14])
	payload := b[14:]

	switch et {
	case 0x8100: // single VLAN tag: strip tag then classify inner payload
		p.Ethertype = "VLAN"
		if len(payload) < 4 {
			return p
		}
		payload = payload[4:]
		classifyInner(p, payload)
		return p
	case 0x0806: // ARP
		p.Ethertype = "ARP"
		p.Proto = "ARP"
		parseARP(p, payload)
		return p
	case 0x86DD:
		p.Ethertype = "IPv6"
		parseIPv6(p, payload)
		return p
	case 0x0800:
		p.Ethertype = "IPv4"
	default:
		p.Ethertype = "Other"
		p.Proto = "Other"
		return p
	}

	if !parseIPv4(p, payload) {
		return nil
	}
	return p
}

func parseIPv4(p *Packet, b []byte) bool {
	if len(b) < 20 || int(b[0]>>4) != 4 {
		return false
	}
	ihl := int(b[0]&0x0f) * 4
	if ihl < 20 || len(b) < ihl {
		return false
	}
	total := int(binary.BigEndian.Uint16(b[2:4]))
	p.SrcIP = ipString(b[12:16])
	p.DstIP = ipString(b[16:20])
	proto := b[9]
	l4 := b[ihl:]
	// slice to claimed length when the capture snap trimmed nothing odd
	if total >= ihl && total <= len(b) {
		l4 = b[ihl:total]
	} else if total > len(b) {
		l4 = b[ihl:] // truncated frame: use what we have
	}
	switch proto {
	case 1:
		p.Proto = "ICMP"
		if len(l4) >= 8 {
			p.PayLen = len(l4) - 8 // fixed 8-byte ICMP header
		}
	case 6:
		return parseTCP(p, l4)
	case 17:
		return parseUDP(p, l4)
	default:
		p.Proto = "IP-" + strconv.Itoa(int(proto))
	}
	return true
}

func parseTCP(p *Packet, b []byte) bool {
	if len(b) < 20 {
		p.Proto = "TCP-trunc"
		return false
	}
	p.Proto = "TCP"
	p.SrcPort = int(binary.BigEndian.Uint16(b[0:2]))
	p.DstPort = int(binary.BigEndian.Uint16(b[2:4]))
	p.Seq = binary.BigEndian.Uint32(b[4:8])
	p.Ack = binary.BigEndian.Uint32(b[8:12])
	dataOff := int(b[12]>>4) * 4
	if dataOff < 20 || dataOff > len(b) {
		dataOff = 20
	}
	flags := b[13]
	p.Flags = renderFlags(flags)
	if len(b) > dataOff {
		p.PayLen = len(b) - dataOff
	}
	p.Win = binary.BigEndian.Uint16(b[14:16])
	return true
}

func parseUDP(p *Packet, b []byte) bool {
	if len(b) < 8 {
		p.Proto = "UDP-trunc"
		return false
	}
	p.Proto = "UDP"
	p.SrcPort = int(binary.BigEndian.Uint16(b[0:2]))
	p.DstPort = int(binary.BigEndian.Uint16(b[2:4]))
	if n := int(binary.BigEndian.Uint16(b[4:6])); n >= 8 && n <= len(b) {
		p.PayLen = n - 8
	}
	return true
}

func parseIPv6(p *Packet, b []byte) {
	p.Proto = "IPv6"
	if len(b) >= 40 {
		p.SrcIP = net.IP(b[8:24]).String()
		p.DstIP = net.IP(b[24:40]).String()
		nh := b[6]
		switch nh {
		case 6:
			p.Proto = "TCP"
			if len(b) > 54 {
				p.Flags = renderFlags(b[53])
			}
		case 17:
			p.Proto = "UDP"
		case 58:
			p.Proto = "ICMPv6"
		}
	}
}

func parseARP(p *Packet, b []byte) {
	// htype(2) ptype(2) hlen plen oper(2) sha(6) spa(4) tha(6) tpa(4)
	if len(b) >= 28 && binary.BigEndian.Uint16(b[2:4]) == 0x0800 {
		p.SrcIP = ipString(b[14:18])
		p.DstIP = ipString(b[24:28])
		op := binary.BigEndian.Uint16(b[6:8])
		if op == 1 {
			p.Flags = "[request]"
		} else {
			p.Flags = "[reply]"
		}
	}
}

// renderFlags renders TCP flag bits tcpdump-style: F S R P . U E C where a
// trailing '.' means ACK is set.
func renderFlags(f uint8) string {
	out := ""
	if f&0x01 != 0 { // FIN
		out += "F"
	}
	if f&0x02 != 0 { // SYN
		out += "S"
	}
	if f&0x04 != 0 { // RST
		out += "R"
	}
	if f&0x08 != 0 { // PSH
		out += "P"
	}
	if f&0x10 != 0 { // ACK rendered as '.'
		out += "."
	}
	if f&0x20 != 0 { // URG
		out += "U"
	}
	if f&0x40 != 0 { // ECE
		out += "E"
	}
	if f&0x80 != 0 { // CWR
		out += "C"
	}
	return "[" + out + "]"
}

func passFilter(p *Packet, host string, port int) bool {
	if host != "" && p.SrcIP != host && p.DstIP != host {
		return false
	}
	if port > 0 {
		hasPorts := p.Proto == "TCP" || p.Proto == "UDP"
		if !hasPorts || (p.SrcPort != port && p.DstPort != port) {
			return false
		}
	}
	return true
}

func collectTCPEvents(res *Result, p *Packet) {
	if p.Proto != "TCP" {
		return
	}
	f := p.Flags
	if containsByte(f, 'S') && len(f) == 3 { // exactly [S]
		res.TCPEvents.Syn++
	}
	if containsByte(f, 'S') && containsByte(f, '.') {
		res.TCPEvents.SynAck++
	}
	if containsByte(f, 'R') {
		res.TCPEvents.Rst++
	}
	if containsByte(f, 'F') {
		res.TCPEvents.Fin++
	}
	if p.Win == 0 && containsByte(f, '.') && !containsByte(f, 'S') && !containsByte(f, 'R') && !containsByte(f, 'F') {
		res.TCPEvents.ZeroWindow++
	}
}

type convCount struct {
	key string
	n   int
}

var convCounts map[string]int

func maybeNoteConversation(res *Result, p *Packet) {
	if convCounts == nil {
		convCounts = map[string]int{}
	}
	if (p.Proto == "TCP" || p.Proto == "UDP") && p.SrcPort > 0 {
		var k string
		if p.SrcPort < p.DstPort {
			k = fmt.Sprintf("%s:%d>%s:%d %s", p.SrcIP, p.SrcPort, p.DstIP, p.DstPort, p.Proto)
		} else {
			k = fmt.Sprintf("%s:%d<%s:%d %s", p.DstIP, p.DstPort, p.SrcIP, p.SrcPort, p.Proto)
		}
		convCounts[k]++
	}
}

func topConversations(res *Result) {
	if len(convCounts) == 0 {
		return
	}
	list := make([]convCount, 0, len(convCounts))
	for k, v := range convCounts {
		list = append(list, convCount{k, v})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	for i, c := range list {
		if i >= 10 {
			break
		}
		res.Conversations = append(res.Conversations,
			fmt.Sprintf("%s x%d", c.key, c.n))
	}
	convCounts = nil // reset for next Capture call in same process
}

// ---------------------------------------------------------------------------
// pcap classic format writer (magic 0xa1b2c3d4, LINKTYPE_ETHERNET=1)
// ---------------------------------------------------------------------------

// WritePcap writes retained frames as a standard libpcap file openable by
// Wireshark / tcpdump / tshark.
func WritePcap(path string, frames []capturedFrame, snaplen int) error {
	if snaplen <= 0 {
		snaplen = 262144
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Classic pcap global header layout:
	// 0 magic | 4 vmaj | 6 vmin | 8 thiszone | 12 sigfigs | 16 snaplen | 20 network
	hdr := make([]byte, 24)
	binary.LittleEndian.PutUint32(hdr[0:4], 0xa1b2c3d4)
	binary.LittleEndian.PutUint16(hdr[4:6], 2)
	binary.LittleEndian.PutUint16(hdr[6:8], 4)
	binary.LittleEndian.PutUint32(hdr[8:12], 0)  // thiszone
	binary.LittleEndian.PutUint32(hdr[12:16], 0) // sigfigs
	binary.LittleEndian.PutUint32(hdr[16:20], uint32(snaplen))
	binary.LittleEndian.PutUint32(hdr[20:24], 1) // LINKTYPE_ETHERNET

	if _, err := f.Write(hdr); err != nil {
		return err
	}
	rec := make([]byte, 16)
	for _, cf := range frames {
		binary.LittleEndian.PutUint32(rec[0:4], uint32(cf.sec))
		binary.LittleEndian.PutUint32(rec[4:8], uint32(cf.usec))
		binary.LittleEndian.PutUint32(rec[8:12], uint32(len(cf.raw)))
		binary.LittleEndian.PutUint32(rec[12:16], uint32(cf.origLen))
		if _, err := f.Write(rec); err != nil {
			return err
		}
		if _, err := f.Write(cf.raw); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// small helpers
// ---------------------------------------------------------------------------

func htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return binary.LittleEndian.Uint16(b)
}

func macString(b []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		b[0], b[1], b[2], b[3], b[4], b[5])
}

func ipString(b []byte) string {
	return net.IP(b).String()
}

// classifyInner labels a frame after a stripped VLAN tag (best effort).
func classifyInner(p *Packet, payload []byte) {
	if len(payload) >= 20 && int(payload[0]>>4) == 4 {
		p.Ethertype = "IPv4"
		if !parseIPv4(p, payload) {
			p.Proto = "Other"
		}
		return
	}
	p.Ethertype = "Other"
	p.Proto = "Other"
}

func containsByte(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
