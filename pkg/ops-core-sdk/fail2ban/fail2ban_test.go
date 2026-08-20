package fail2ban

import "testing"

func TestJailStatusEmpty(t *testing.T) {
	_, err := JailStatus("")
	if err == nil {
		t.Error("expected error for empty jail name")
	}
}

func TestBanIPEmptyJail(t *testing.T) {
	_, err := BanIP("", "1.2.3.4")
	if err == nil {
		t.Error("expected error for empty jail name")
	}
}

func TestBanIPEmptyIP(t *testing.T) {
	_, err := BanIP("sshd", "")
	if err == nil {
		t.Error("expected error for empty IP")
	}
}

func TestUnbanIPEmptyJail(t *testing.T) {
	_, err := UnbanIP("", "1.2.3.4")
	if err == nil {
		t.Error("expected error for empty jail name")
	}
}

func TestUnbanIPEmptyIP(t *testing.T) {
	_, err := UnbanIP("sshd", "")
	if err == nil {
		t.Error("expected error for empty IP")
	}
}

func TestGet(t *testing.T) {
	// Will return Running=false if fail2ban not installed, which is OK
	result, _ := Get()
	if result == nil {
		t.Error("expected non-nil result")
	}
}
