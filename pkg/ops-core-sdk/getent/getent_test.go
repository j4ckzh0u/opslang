package getent

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLookupUserEmpty(t *testing.T) {
	_, err := LookupUser("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLookupGroupEmpty(t *testing.T) {
	_, err := LookupGroup("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLookupServiceEmpty(t *testing.T) {
	_, err := LookupService("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLookupProtocolEmpty(t *testing.T) {
	_, err := LookupProtocol("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLookupRoot(t *testing.T) {
	res, err := LookupUser("root")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected root to be found")
	}
	if res.Database != "passwd" {
		t.Fatalf("expected database 'passwd', got %q", res.Database)
	}
}

func TestLookupRootUID(t *testing.T) {
	res, err := LookupUser("0")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected UID 0 to be found")
	}
}

func TestLookupGroupByGID(t *testing.T) {
	res, err := LookupGroup("0")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected GID 0 to be found")
	}
}

func TestLookupNonExistentUser(t *testing.T) {
	res, err := LookupUser("nonexistent_user_xyz_12345")
	if err != nil {
		t.Fatal(err)
	}
	if res.Found {
		t.Fatal("expected not found")
	}
}

func TestLookupService(t *testing.T) {
	if _, err := os.Stat("/etc/services"); err != nil {
		t.Skipf("/etc/services unavailable: %v", err)
	}
	res, err := LookupService("ssh")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected ssh service to be found")
	}
}

func TestLookupProtocol(t *testing.T) {
	if _, err := os.Stat("/etc/protocols"); err != nil {
		t.Skipf("/etc/protocols unavailable: %v", err)
	}
	res, err := LookupProtocol("tcp")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Found {
		t.Fatal("expected tcp protocol to be found")
	}
}

func TestGetPasswd(t *testing.T) {
	entries, err := GetPasswd()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one user")
	}
	found := false
	for _, e := range entries {
		if e.Username == "root" {
			found = true
			if e.UID != 0 {
				t.Fatalf("expected root UID 0, got %d", e.UID)
			}
		}
	}
	if !found {
		t.Fatal("root not found in passwd")
	}
}

func TestGetGroups(t *testing.T) {
	entries, err := GetGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one group")
	}
}

func TestShells(t *testing.T) {
	res, err := Shells()
	if err != nil {
		// /etc/shells may not exist on all systems, that's ok
		return
	}
	if res.Count < 0 {
		t.Fatal("unexpected negative count")
	}
}

func TestLookupResultJSON(t *testing.T) {
	r := LookupResult{Database: "passwd", Key: "root", Found: true, Count: 1}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LookupResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Found || decoded.Count != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestUserEntryJSON(t *testing.T) {
	e := UserEntry{Username: "root", UID: 0, GID: 0, Home: "/root", Shell: "/bin/bash"}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UserEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Username != "root" || decoded.UID != 0 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}

func TestGroupEntryJSON(t *testing.T) {
	e := GroupEntry{Name: "wheel", GID: 10, Members: []string{"root"}}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GroupEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != "wheel" || len(decoded.Members) != 1 {
		t.Fatalf("unexpected: %+v", decoded)
	}
}
