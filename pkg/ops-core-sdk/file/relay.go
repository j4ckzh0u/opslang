package file

import (
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

const (
	defaultRelayThreshold  = 20
	defaultRelayMaxTargets = 100
)

// RelayPlan is a deterministic distribution topology.
type RelayPlan struct {
	Groups []RelayGroupPlan `json:"groups"`
}

// RelayGroupPlan describes a relay fan-out or a direct-transfer set.
type RelayGroupPlan struct {
	Key     string             `json:"key"`
	Relay   *DistributeTarget  `json:"relay,omitempty"`
	Targets []DistributeTarget `json:"targets,omitempty"`
	Direct  []DistributeTarget `json:"direct,omitempty"`
}

// BuildRelayPlan groups targets by explicit metadata or network prefix.
func BuildRelayPlan(targets []DistributeTarget, opts DistributeOptions) (RelayPlan, error) {
	threshold := opts.RelayThreshold
	if threshold == 0 {
		threshold = defaultRelayThreshold
	}
	maxTargets := opts.RelayMaxTargets
	if maxTargets == 0 {
		maxTargets = defaultRelayMaxTargets
	}
	if threshold < 1 {
		return RelayPlan{}, fmt.Errorf("relay_threshold must be at least 1")
	}
	if maxTargets < 1 {
		return RelayPlan{}, fmt.Errorf("relay_max_targets must be at least 1")
	}

	grouped := make(map[string][]DistributeTarget)
	for index, target := range targets {
		if strings.TrimSpace(target.Host) == "" {
			return RelayPlan{}, fmt.Errorf("target %d host is empty", index)
		}
		key, direct := relayTopologyKey(target, opts.RelayGroup)
		if direct {
			key = "direct:" + relayTargetIdentity(target)
		}
		grouped[key] = append(grouped[key], cloneDistributeTarget(target))
	}

	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	plan := RelayPlan{Groups: make([]RelayGroupPlan, 0, len(keys))}
	for _, key := range keys {
		members := grouped[key]
		sort.SliceStable(members, func(i, j int) bool {
			leftPreferred := strings.EqualFold(members[i].Tags["relay"], "true")
			rightPreferred := strings.EqualFold(members[j].Tags["relay"], "true")
			if leftPreferred != rightPreferred {
				return leftPreferred
			}
			return relayTargetIdentity(members[i]) < relayTargetIdentity(members[j])
		})
		if !opts.Relay || strings.HasPrefix(key, "direct:") || len(members) < threshold {
			plan.Groups = append(plan.Groups, RelayGroupPlan{Key: key, Direct: members})
			continue
		}
		chunkSize := maxTargets + 1
		for start, part := 0, 0; start < len(members); part++ {
			end := min(start+chunkSize, len(members))
			chunk := members[start:end]
			if len(chunk) == 1 {
				plan.Groups = append(plan.Groups, RelayGroupPlan{Key: relayPartKey(key, part), Direct: chunk})
			} else {
				relay := cloneDistributeTarget(chunk[0])
				plan.Groups = append(plan.Groups, RelayGroupPlan{
					Key:     relayPartKey(key, part),
					Relay:   &relay,
					Targets: append([]DistributeTarget(nil), chunk[1:]...),
				})
			}
			start = end
		}
	}
	return plan, nil
}

func relayTopologyKey(target DistributeTarget, defaultGroup string) (string, bool) {
	if value := strings.TrimSpace(target.RelayGroup); value != "" {
		return "group:" + value, false
	}
	if value := strings.TrimSpace(target.Tags["relay_group"]); value != "" {
		return "group:" + value, false
	}
	if value := strings.TrimSpace(defaultGroup); value != "" {
		return "group:" + value, false
	}
	address, err := netip.ParseAddr(strings.TrimSpace(target.Host))
	if err != nil {
		return "", true
	}
	bits := 64
	if address.Is4() {
		bits = 24
	}
	return "network:" + netip.PrefixFrom(address, bits).Masked().String(), false
}

func relayPartKey(key string, part int) string {
	return key + ":part:" + strconv.Itoa(part)
}

func relayTargetIdentity(target DistributeTarget) string {
	return target.Host + ":" + strconv.Itoa(target.Port) + ":" + target.User + ":" + target.Dest
}

func cloneDistributeTarget(target DistributeTarget) DistributeTarget {
	clone := target
	if target.Tags != nil {
		clone.Tags = make(map[string]string, len(target.Tags))
		for key, value := range target.Tags {
			clone.Tags[key] = value
		}
	}
	return clone
}
