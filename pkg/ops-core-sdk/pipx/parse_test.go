package pipx

import (
	"reflect"
	"testing"
)

// TestParseListOutput pins the list parsing without needing a real pipx:
// CI and dev machines run different pipx versions whose --short output
// differs (name-only vs name+version), and the whole-line comparison used
// to make idempotent installs report Changed forever.
func TestParseListOutput(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "old format, one name per line",
			in:   "cowsay\nblack\n",
			want: []string{"cowsay", "black"},
		},
		{
			name: "new format, name plus version",
			in:   "cowsay 6.0\nblack 24.3.0\n",
			want: []string{"cowsay", "black"},
		},
		{
			name: "mixed formats and blank lines",
			in:   "cowsay 6.0\n\nblack\n  ruff 0.3  \n",
			want: []string{"cowsay", "black", "ruff"},
		},
		{
			name: "empty output",
			in:   "",
			want: nil,
		},
	}
	for _, tc := range cases {
		got := parseListOutput(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: parseListOutput(%q) = %#v, want %#v", tc.name, tc.in, got, tc.want)
		}
	}
}
