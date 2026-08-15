package relay

import "fmt"

// BuildTopology constructs a hierarchical relay tree from a flat list of hosts.
// Hosts are grouped by /24 subnet. Groups with >= MinHostsForRelay hosts get
// relay nodes selected. The root node (Tier 0) represents the controller.
func BuildTopology(hosts []HostInfo, cfg Config) (*RelayNode, error) {
	cfg.Defaults()

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts provided")
	}

	// Group hosts by subnet.
	groups := make(map[string][]HostInfo)
	for _, h := range hosts {
		subnet := extractSubnet(h.Host)
		groups[subnet] = append(groups[subnet], h)
	}

	root := &RelayNode{
		Host:     "controller",
		Port:     0,
		User:     "",
		Tier:     0,
		Capacity: len(hosts),
	}

	for subnet, group := range groups {
		if len(group) < cfg.MinHostsForRelay {
			// Small group: attach directly as leaves.
			for _, h := range group {
				hCopy := h
				child := &RelayNode{
					Host:     h.Host,
					Port:     h.Port,
					User:     h.User,
					Tier:     1,
					Subnet:   subnet,
					IsLeaf:   true,
					HostInfo: &hCopy,
				}
				root.Children = append(root.Children, child)
			}
			continue
		}

		// Large group: select relay nodes.
		// Determine number of relays: roughly 1 per 50 hosts, at least 1.
		relayCount := len(group) / 50
		if relayCount < 1 {
			relayCount = 1
		}

		relays := SelectRelays(group, relayCount)
		relaySet := make(map[string]bool)
		for _, r := range relays {
			relaySet[r.Host] = true
		}

		// Determine if we need L2 relays (>= 100 hosts and RelayDepth >= 2).
		needL2 := len(group) >= 100 && cfg.RelayDepth >= 2

		if needL2 && len(relays) > 1 {
			buildL2Topology(root, subnet, group, relays, relaySet)
		} else {
			buildL1Topology(root, subnet, group, relays, relaySet)
		}
	}

	return root, nil
}

// buildL1Topology builds a single-level relay tree: L1 relays with leaves beneath them.
func buildL1Topology(root *RelayNode, subnet string, group []HostInfo, relays []HostInfo, relaySet map[string]bool) {
	// Build relay node slice so we can round-robin leaves across them.
	relayNodes := make([]*RelayNode, 0, len(relays))
	perRelayCapacity := len(group) / len(relays)
	for _, r := range relays {
		rCopy := r
		relayNode := &RelayNode{
			Host:     r.Host,
			Port:     r.Port,
			User:     r.User,
			Tier:     1,
			Subnet:   subnet,
			Capacity: perRelayCapacity,
			HostInfo: &rCopy,
		}
		root.Children = append(root.Children, relayNode)
		relayNodes = append(relayNodes, relayNode)
	}

	// Attach non-relay hosts as leaves under relays (round-robin).
	leafIdx := 0
	for _, h := range group {
		if relaySet[h.Host] {
			continue
		}
		hCopy := h
		leaf := &RelayNode{
			Host:     h.Host,
			Port:     h.Port,
			User:     h.User,
			Tier:     2,
			Subnet:   subnet,
			IsLeaf:   true,
			HostInfo: &hCopy,
		}
		relayNodes[leafIdx%len(relayNodes)].Children = append(relayNodes[leafIdx%len(relayNodes)].Children, leaf)
		leafIdx++
	}
}

// buildL2Topology builds a two-level relay tree: L1 relay -> L2 relays -> leaves.
func buildL2Topology(root *RelayNode, subnet string, group []HostInfo, relays []HostInfo, relaySet map[string]bool) {
	l1Relay := relays[0]
	l1Node := &RelayNode{
		Host:     l1Relay.Host,
		Port:     l1Relay.Port,
		User:     l1Relay.User,
		Tier:     1,
		Subnet:   subnet,
		Capacity: len(group),
		Metadata: map[string]string{"role": "l1-relay"},
	}

	l2Relays := relays[1:]
	l2Nodes := make([]*RelayNode, 0, len(l2Relays))
	perL2Capacity := len(group) / len(l2Relays)
	for _, l2r := range l2Relays {
		l2Node := &RelayNode{
			Host:     l2r.Host,
			Port:     l2r.Port,
			User:     l2r.User,
			Tier:     2,
			Subnet:   subnet,
			Capacity: perL2Capacity,
			Metadata: map[string]string{"role": "l2-relay"},
		}
		l1Node.Children = append(l1Node.Children, l2Node)
		l2Nodes = append(l2Nodes, l2Node)
	}

	// Attach non-relay hosts as leaves under L2 relays (round-robin).
	leafIdx := 0
	for _, h := range group {
		if relaySet[h.Host] {
			continue
		}
		hCopy := h
		leaf := &RelayNode{
			Host:     h.Host,
			Port:     h.Port,
			User:     h.User,
			Tier:     3,
			Subnet:   subnet,
			IsLeaf:   true,
			HostInfo: &hCopy,
		}
		l2Node := l2Nodes[leafIdx%len(l2Nodes)]
		l2Node.Children = append(l2Node.Children, leaf)
		leafIdx++
	}

	root.Children = append(root.Children, l1Node)
}

// extractSubnet extracts the /24 prefix from an IP address.
// For "10.0.1.5" returns "10.0.1". For hostnames, returns as-is.
func extractSubnet(host string) string {
	parts := 0
	for i := 0; i < len(host); i++ {
		if host[i] == '.' {
			parts++
			if parts == 3 {
				return host[:i]
			}
		}
	}
	// Not an IPv4 address with 4 parts.
	return host
}
