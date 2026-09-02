package sys

import "testing"

func TestCalculateNetRate(t *testing.T) {
	before := map[string]NetIOCounters{
		"eth0": {Name: "eth0", BytesSent: 1_000, BytesRecv: 2_000},
		"lo":   {Name: "lo", BytesSent: 100, BytesRecv: 200},
	}
	after := map[string]NetIOCounters{
		"eth0": {Name: "eth0", BytesSent: 4_000, BytesRecv: 6_000},
		"lo":   {Name: "lo", BytesSent: 500, BytesRecv: 800},
	}
	got := calculateNetRate(before, after, 3)
	if got.WindowSeconds != 3 || got.BytesSent != 3_400 || got.BytesRecv != 4_600 {
		t.Fatalf("unexpected aggregate: %+v", got)
	}
	if got.BitsPerSecond != (8_000*8.0)/3.0 {
		t.Fatalf("bits/sec = %v", got.BitsPerSecond)
	}
	if len(got.Interfaces) != 2 {
		t.Fatalf("interfaces = %d, want 2", len(got.Interfaces))
	}
}

func TestCalculateNetRateCounterReset(t *testing.T) {
	before := map[string]NetIOCounters{"eth0": {Name: "eth0", BytesSent: 100, BytesRecv: 100}}
	after := map[string]NetIOCounters{"eth0": {Name: "eth0", BytesSent: 10, BytesRecv: 20}}
	got := calculateNetRate(before, after, 3)
	if got.BytesSent != 0 || got.BytesRecv != 0 || got.BitsPerSecond != 0 {
		t.Fatalf("counter reset must not produce negative rate: %+v", got)
	}
}
