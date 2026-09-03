package file

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestBuildRelayPlanTopologyPriority(t *testing.T) {
	targets := []DistributeTarget{
		{Host: "10.0.1.10", RelayGroup: "explicit", Tags: map[string]string{"relay_group": "tagged"}},
		{Host: "10.0.2.10", Tags: map[string]string{"relay_group": "tagged"}},
		{Host: "10.0.3.10"},
		{Host: "10.0.3.99"},
		{Host: "2001:db8:1::1"},
		{Host: "2001:db8:1::2"},
		{Host: "hostname.example"},
	}
	plan, err := BuildRelayPlan(targets, DistributeOptions{Relay: true, RelayThreshold: 2})
	if err != nil {
		t.Fatalf("BuildRelayPlan: %v", err)
	}
	keys := make(map[string]RelayGroupPlan, len(plan.Groups))
	for _, group := range plan.Groups {
		keys[group.Key] = group
	}
	for _, key := range []string{"group:explicit", "group:tagged", "network:10.0.3.0/24:part:0", "network:2001:db8:1::/64:part:0", "direct:hostname.example:0::"} {
		if _, ok := keys[key]; !ok {
			t.Errorf("missing group %q in %+v", key, plan.Groups)
		}
	}
	if keys["group:explicit"].Relay != nil {
		t.Fatal("single explicit group member must use direct transfer")
	}
	if keys["network:10.0.3.0/24:part:0"].Relay == nil {
		t.Fatal("IPv4 /24 group must use a relay")
	}
}

func TestBuildRelayPlanDeterministicAndConservative(t *testing.T) {
	targets := make([]DistributeTarget, 0, 205)
	for index := 0; index < 205; index++ {
		target := DistributeTarget{Host: "host-" + padIndex(index), RelayGroup: "rack-a"}
		if index == 17 {
			target.Tags = map[string]string{"relay": "true"}
		}
		targets = append(targets, target)
	}
	want, err := BuildRelayPlan(targets, DistributeOptions{Relay: true, RelayThreshold: 20, RelayMaxTargets: 100})
	if err != nil {
		t.Fatalf("BuildRelayPlan: %v", err)
	}
	shuffled := append([]DistributeTarget(nil), targets...)
	rand.New(rand.NewSource(42)).Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	got, err := BuildRelayPlan(shuffled, DistributeOptions{Relay: true, RelayThreshold: 20, RelayMaxTargets: 100})
	if err != nil {
		t.Fatalf("BuildRelayPlan shuffled: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("same target set produced a different relay plan after shuffling")
	}
	seen := make(map[string]int, len(targets))
	for _, group := range got.Groups {
		if len(group.Targets) > 100 {
			t.Fatalf("group %s has %d relay targets", group.Key, len(group.Targets))
		}
		if group.Relay != nil {
			seen[group.Relay.Host]++
		}
		for _, target := range group.Targets {
			seen[target.Host]++
		}
		for _, target := range group.Direct {
			seen[target.Host]++
		}
	}
	if len(seen) != len(targets) {
		t.Fatalf("planned hosts = %d, want %d", len(seen), len(targets))
	}
	for host, count := range seen {
		if count != 1 {
			t.Fatalf("host %s appears %d times", host, count)
		}
	}
	if got.Groups[0].Relay == nil || got.Groups[0].Relay.Host != "host-017" {
		t.Fatalf("preferred relay = %+v, want host-017", got.Groups[0].Relay)
	}
}

func TestBuildRelayPlanDefaultsAndValidation(t *testing.T) {
	plan, err := BuildRelayPlan([]DistributeTarget{{Host: "10.0.0.1"}}, DistributeOptions{Relay: true})
	if err != nil {
		t.Fatalf("BuildRelayPlan defaults: %v", err)
	}
	if len(plan.Groups) != 1 || len(plan.Groups[0].Direct) != 1 {
		t.Fatalf("small default group = %+v", plan)
	}
	if _, err := BuildRelayPlan(nil, DistributeOptions{RelayThreshold: -1}); err == nil {
		t.Fatal("negative relay threshold must fail")
	}
	if _, err := BuildRelayPlan(nil, DistributeOptions{RelayMaxTargets: -1}); err == nil {
		t.Fatal("negative relay max targets must fail")
	}
	if _, err := BuildRelayPlan([]DistributeTarget{{}}, DistributeOptions{}); err == nil {
		t.Fatal("empty host must fail")
	}
}

func padIndex(index int) string {
	return fmt.Sprintf("%03d", index)
}
