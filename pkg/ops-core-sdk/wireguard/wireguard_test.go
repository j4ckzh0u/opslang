package wireguard

import "testing"

func TestUpEmpty(t *testing.T) {
	_, err := Up("", "")
	if err == nil {
		t.Error("expected error for empty interface name")
	}
}

func TestDownEmpty(t *testing.T) {
	_, err := Down("")
	if err == nil {
		t.Error("expected error for empty interface name")
	}
}

func TestAddPeerEmptyInterface(t *testing.T) {
	_, err := AddPeer("", "key", "", "")
	if err == nil {
		t.Error("expected error for empty interface name")
	}
}

func TestAddPeerEmptyKey(t *testing.T) {
	_, err := AddPeer("wg0", "", "", "")
	if err == nil {
		t.Error("expected error for empty public key")
	}
}

func TestRemovePeerEmptyInterface(t *testing.T) {
	_, err := RemovePeer("", "key")
	if err == nil {
		t.Error("expected error for empty interface name")
	}
}

func TestRemovePeerEmptyKey(t *testing.T) {
	_, err := RemovePeer("wg0", "")
	if err == nil {
		t.Error("expected error for empty public key")
	}
}

func TestPubKeyEmpty(t *testing.T) {
	_, err := PubKey("")
	if err == nil {
		t.Error("expected error for empty private key")
	}
}

func TestShow(t *testing.T) {
	// Show will return empty if wg not installed
	result, _ := Show()
	if result == nil {
		t.Error("expected non-nil result")
	}
}
