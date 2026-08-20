package docker_image

import "testing"

func TestPullValidation(t *testing.T) {
	_, err := Pull("", "", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestBuildValidation(t *testing.T) {
	_, err := Build("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty path/name")
	}
}

func TestRemoveValidation(t *testing.T) {
	_, err := Remove("", "", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestTagValidation(t *testing.T) {
	_, err := Tag("", "target")
	if err == nil {
		t.Fatal("expected error for empty source")
	}
	_, err = Tag("source", "")
	if err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestInspectValidation(t *testing.T) {
	_, err := Inspect("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestPushValidation(t *testing.T) {
	_, err := Push("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestFindDocker(t *testing.T) {
	_ = findDocker()
}
