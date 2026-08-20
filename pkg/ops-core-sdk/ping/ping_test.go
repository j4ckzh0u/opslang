package ping

import "testing"

func TestPingDefault(t *testing.T) {
	r := Ping("")
	if r.Ping != "pong" {
		t.Errorf("expected pong, got %s", r.Ping)
	}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestPingData(t *testing.T) {
	r := Ping("hello")
	if r.Ping != "hello" {
		t.Errorf("expected hello, got %s", r.Ping)
	}
}

func TestWinPing(t *testing.T) {
	r := WinPing("")
	if r.Ping != "pong" {
		t.Error("expected pong")
	}
}
