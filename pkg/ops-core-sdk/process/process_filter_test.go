package process

import "testing"

// TestIsKernelThread pins the kernel-thread name rule.
func TestIsKernelThread(t *testing.T) {
	cases := []struct {
		desc string
		name string
		want bool
	}{
		{"kthreadd bracketed", "[kthreadd]", true},
		{"kworker", "[kworker/u8:2]", true},
		{"softirq", "[ksoftirqd/0]", true},
		{"plain nginx", "nginx", false},
		{"java with dots", "java", false},
		{"empty name edge", "", false},
		{"only open bracket", "[weird", false},
		{"only close bracket", "weird]", false},
	}
	for _, tc := range cases {
		if got := IsKernelThread(tc.name); got != tc.want {
			t.Errorf("%s: IsKernelThread(%q) = %v, want %v", tc.desc, tc.name, got, tc.want)
		}
	}
}

// TestListNoKernelThreads runs the real call and asserts no bracketed
// kernel-thread names leak into the result.
func TestListNoKernelThreads(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	for _, p := range procs {
		if IsKernelThread(p.Name) {
			t.Errorf("kernel thread leaked into List(): pid=%d name=%q", p.PID, p.Name)
		}
	}
}
