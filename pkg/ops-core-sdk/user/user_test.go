package user

import (
	"os"
	"testing"
)

// isRoot returns true when the test process is running as uid 0.
func isRoot() bool {
	return os.Getuid() == 0
}

// existingUsername returns a username that is guaranteed to exist on any Linux box.
func existingUsername() string {
	return "root"
}

func TestInfo(t *testing.T) {
	username := existingUsername()

	result, err := Info(username)
	if err != nil {
		t.Fatalf("Info(%q) unexpected error: %v", username, err)
	}
	info := result.User

	tests := []struct {
		name  string
		check func() bool
	}{
		{"UID is 0 for root", func() bool { return info.UID == 0 }},
		{"GID is 0 for root", func() bool { return info.GID == 0 }},
		{"Username matches", func() bool { return info.Username == username }},
		{"Home is not empty", func() bool { return info.Home != "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("Info(%q) check %q failed: %+v", username, tt.name, info)
			}
		})
	}
}

func TestInfo_NotFound(t *testing.T) {
	_, err := Info("__opslang_nonexistent_user_xyz__")
	if err == nil {
		t.Error("Info(nonexistent) expected error, got nil")
	}
}

func TestList(t *testing.T) {
	result, err := List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}
	if len(result.Users) == 0 {
		t.Fatal("List() returned zero users")
	}

	// Root must appear.
	found := false
	for _, u := range result.Users {
		if u.Username == "root" {
			found = true
			if u.UID != 0 {
				t.Errorf("root UID expected 0, got %d", u.UID)
			}
			break
		}
	}
	if !found {
		t.Error("List() did not contain root")
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"existing user", existingUsername(), true},
		{"non-existing user", "__opslang_nonexistent_user_xyz__", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Exists(tt.username)
			if err != nil {
				t.Fatalf("Exists(%q) unexpected error: %v", tt.username, err)
			}
			if res.Exists != tt.want {
				t.Errorf("Exists(%q).Exists = %v, want %v", tt.username, res.Exists, tt.want)
			}
		})
	}
}

// --- Tests below require root. They are skipped automatically otherwise. ---

func TestAdd_RequiresRoot(t *testing.T) {
	if !isRoot() {
		t.Skip("skipping: requires root")
	}
	username := "__opslang_test_user_add__"
	// Clean up in case of prior failure.
	_, _ = Remove(username, true)

	res, err := Add(username, map[string]string{
		"shell":       "/bin/sh",
		"create_home": "false",
	})
	if err != nil {
		t.Fatalf("Add(%q) error: %v", username, err)
	}
	if !res.Changed {
		t.Errorf("Add(%q).Changed = false, want true; error=%q", username, res.Error)
	}
	if res.UID == 0 {
		t.Errorf("Add(%q).UID = 0, want non-zero", username)
	}

	// Verify existence.
	ex, _ := Exists(username)
	if !ex.Exists {
		t.Errorf("user %q should exist after Add", username)
	}

	// Clean up.
	_, _ = Remove(username, true)
}

func TestRemove_RequiresRoot(t *testing.T) {
	if !isRoot() {
		t.Skip("skipping: requires root")
	}
	username := "__opslang_test_user_rm__"
	_, _ = Remove(username, true) // ensure clean state
	_, err := Add(username, map[string]string{"create_home": "false"})
	if err != nil {
		t.Fatalf("setup Add(%q) error: %v", username, err)
	}

	res, err := Remove(username, false)
	if err != nil {
		t.Fatalf("Remove(%q) error: %v", username, err)
	}
	if !res.Changed {
		t.Errorf("Remove(%q).Changed = false; error=%q", username, res.Error)
	}

	// Removing a non-existent user should yield Changed=false, no error.
	res2, err2 := Remove("__opslang_nonexistent__", false)
	if err2 != nil {
		t.Fatalf("Remove(nonexistent) unexpected error: %v", err2)
	}
	if res2.Changed {
		t.Error("Remove(nonexistent).Changed = true, want false")
	}
}

func TestModify_RequiresRoot(t *testing.T) {
	if !isRoot() {
		t.Skip("skipping: requires root")
	}
	username := "__opslang_test_user_mod__"
	_, _ = Remove(username, true)
	_, err := Add(username, map[string]string{
		"shell":       "/bin/sh",
		"create_home": "false",
	})
	if err != nil {
		t.Fatalf("setup Add(%q) error: %v", username, err)
	}
	defer func() { _, _ = Remove(username, true) }()

	// Change shell.
	res, err := Modify(username, map[string]string{"shell": "/bin/bash"})
	if err != nil {
		t.Fatalf("Modify(%q) error: %v", username, err)
	}
	if !res.Changed {
		t.Errorf("Modify(%q).Changed = false; error=%q", username, res.Error)
	}

	// No-op modify.
	res2, err2 := Modify(username, map[string]string{})
	if err2 != nil {
		t.Fatalf("Modify(no-op) unexpected error: %v", err2)
	}
	if res2.Changed {
		t.Error("Modify(no-op).Changed = true, want false")
	}

	// Non-existent user.
	res3, err3 := Modify("__opslang_nonexistent__", map[string]string{"shell": "/bin/sh"})
	if err3 == nil {
		t.Error("Modify(nonexistent) expected error, got nil")
	}
	if res3.Changed {
		t.Error("Modify(nonexistent).Changed = true, want false")
	}
}
