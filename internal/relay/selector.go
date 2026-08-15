package relay

import "sort"

// SelectRelays selects the top `count` hosts to serve as relay nodes based on scoring.
func SelectRelays(hosts []HostInfo, count int) []HostInfo {
	if count <= 0 || len(hosts) == 0 {
		return nil
	}
	if count >= len(hosts) {
		result := make([]HostInfo, len(hosts))
		copy(result, hosts)
		return result
	}

	type scored struct {
		host  HostInfo
		score int
	}

	scoredHosts := make([]scored, len(hosts))
	for i, h := range hosts {
		scoredHosts[i] = scored{host: h, score: ScoreNode(h)}
	}

	sort.Slice(scoredHosts, func(i, j int) bool {
		return scoredHosts[i].score > scoredHosts[j].score
	})

	result := make([]HostInfo, count)
	for i := 0; i < count; i++ {
		result[i] = scoredHosts[i].host
	}
	return result
}

// ScoreNode computes a suitability score for a host to be a relay node.
// Higher scores indicate better suitability for relay duties.
func ScoreNode(h HostInfo) int {
	score := 10 // base score

	if h.Tags != nil {
		if h.Tags["relay"] == "true" {
			score += 1000
		}
		if h.Tags["role"] == "relay" {
			score += 500
		}
	}

	if h.Password != "" || h.KeyFile != "" {
		score += 100
	}

	return score
}
