package add_host

import "testing"

func TestAddHostEmpty(t *testing.T) {
	r := Add("", nil, nil)
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}

func TestAddHostAndGet(t *testing.T) {
	// reset
	mu.Lock()
	hosts = map[string]map[string]string{}
	groups = map[string][]string{}
	mu.Unlock()

	r := Add("host1", []string{"web"}, map[string]string{"ansible_host": "10.0.0.1"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}

	v, ok := GetHost("host1")
	if !ok {
		t.Error("expected host to exist")
	}
	if v["ansible_host"] != "10.0.0.1" {
		t.Errorf("unexpected var: %v", v)
	}

	gh := GetGroup("web")
	if len(gh) != 1 || gh[0] != "host1" {
		t.Errorf("unexpected group hosts: %v", gh)
	}

	all := ListHosts()
	if len(all) != 1 {
		t.Errorf("expected 1 host, got %d", len(all))
	}
}
