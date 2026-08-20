package group_by

import "testing"

func TestGroupByEmptyKey(t *testing.T) {
	r := GroupBy(nil, "")
	if r.Error == "" {
		t.Error("expected error for empty key")
	}
}

func TestGroupByAndGet(t *testing.T) {
	Clear()
	r := GroupBy([]string{"host1", "host2"}, "web_servers")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}

	hosts := GetHosts("web_servers")
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestGroupByDedup(t *testing.T) {
	Clear()
	GroupBy([]string{"host1"}, "grp")
	GroupBy([]string{"host1", "host2"}, "grp")

	hosts := GetHosts("grp")
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts (dedup), got %d", len(hosts))
	}
}

func TestListGroups(t *testing.T) {
	Clear()
	GroupBy([]string{"h1"}, "a")
	GroupBy([]string{"h2"}, "b")

	gs := ListGroups()
	if len(gs) != 2 {
		t.Errorf("expected 2 groups, got %d", len(gs))
	}
}
