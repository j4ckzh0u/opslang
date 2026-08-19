package motd

import "testing"

func TestRead(t *testing.T) {
	r := Read()
	// May fail if /etc/motd doesn't exist
	_ = r.Content
}
