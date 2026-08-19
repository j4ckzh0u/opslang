package issue

import "testing"

func TestRead(t *testing.T) {
	r := Read()
	// May fail if /etc/issue doesn't exist
	_ = r.Content
}
