package security

import (
	"testing"
)

func TestDefaultResourceLimits(t *testing.T) {
	limits := DefaultResourceLimits()
	if limits == nil {
		t.Fatal("DefaultResourceLimits returned nil")
	}
	if limits.CPUPercent != 100 {
		t.Errorf("CPUPercent = %d, want 100", limits.CPUPercent)
	}
	if limits.MemoryMB != 1024 {
		t.Errorf("MemoryMB = %d, want 1024", limits.MemoryMB)
	}
	if limits.CPUQuota != "100%" {
		t.Errorf("CPUQuota = %q, want %q", limits.CPUQuota, "100%")
	}
	if limits.MemoryLimit != "1024M" {
		t.Errorf("MemoryLimit = %q, want %q", limits.MemoryLimit, "1024M")
	}
}

func TestApplyResourceLimitsNil(t *testing.T) {
	if err := ApplyResourceLimits(nil); err != nil {
		t.Errorf("ApplyResourceLimits(nil) = %v, want nil", err)
	}
}

func TestApplyResourceLimitsUlimitFallback(t *testing.T) {
	limits := &ResourceLimits{
		CPUPercent:  50,
		MemoryMB:    256,
		CPUQuota:    "50%",
		MemoryLimit: "256M",
	}
	// On non-linux or without systemd, this should use ulimit path (no error expected)
	err := ApplyResourceLimits(limits)
	if err != nil {
		t.Errorf("ApplyResourceLimits() = %v, want nil", err)
	}
}

func TestApplyResourceLimitsZeroMemory(t *testing.T) {
	limits := &ResourceLimits{
		CPUPercent: 50,
		MemoryMB:   0, // skip memory limit
	}
	err := ApplyResourceLimits(limits)
	if err != nil {
		t.Errorf("ApplyResourceLimits() = %v", err)
	}
}

func TestParseResourceLimits(t *testing.T) {
	tests := []struct {
		name       string
		cpu        string
		memory     string
		wantErr    bool
		wantCPU    int
		wantMem    int
		wantQuota  string
		wantMemLim string
	}{
		{
			name: "both empty uses defaults",
			cpu:  "", memory: "",
			wantCPU: 100, wantMem: 1024,
			wantQuota: "100%", wantMemLim: "1024M",
		},
		{
			name: "cpu only",
			cpu:  "50%", memory: "",
			wantCPU: 50, wantMem: 1024,
			wantQuota: "50%", wantMemLim: "1024M",
		},
		{
			name: "memory in MB",
			cpu:  "", memory: "512M",
			wantCPU: 100, wantMem: 512,
			wantQuota: "100%", wantMemLim: "512M",
		},
		{
			name: "memory in GB",
			cpu:  "", memory: "2G",
			wantCPU: 100, wantMem: 2048,
			wantQuota: "100%", wantMemLim: "2048M",
		},
		{
			name: "memory in KB",
			cpu:  "", memory: "1024K",
			wantCPU: 100, wantMem: 1,
			wantQuota: "100%", wantMemLim: "1M",
		},
		{
			name: "memory plain number (assumed MB)",
			cpu:  "", memory: "256",
			wantCPU: 100, wantMem: 256,
			wantQuota: "100%", wantMemLim: "256M",
		},
		{
			name: "both set",
			cpu:  "200%", memory: "1G",
			wantCPU: 200, wantMem: 1024,
			wantQuota: "200%", wantMemLim: "1024M",
		},
		{
			name:    "invalid cpu",
			cpu:     "abc%",
			wantErr: true,
		},
		{
			name:    "cpu too low",
			cpu:     "0%",
			wantErr: true,
		},
		{
			name:    "cpu too high",
			cpu:     "1001%",
			wantErr: true,
		},
		{
			name:    "invalid memory",
			memory:  "abcM",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits, err := ParseResourceLimits(tt.cpu, tt.memory)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if limits.CPUPercent != tt.wantCPU {
				t.Errorf("CPUPercent = %d, want %d", limits.CPUPercent, tt.wantCPU)
			}
			if limits.MemoryMB != tt.wantMem {
				t.Errorf("MemoryMB = %d, want %d", limits.MemoryMB, tt.wantMem)
			}
			if limits.CPUQuota != tt.wantQuota {
				t.Errorf("CPUQuota = %q, want %q", limits.CPUQuota, tt.wantQuota)
			}
			if limits.MemoryLimit != tt.wantMemLim {
				t.Errorf("MemoryLimit = %q, want %q", limits.MemoryLimit, tt.wantMemLim)
			}
		})
	}
}

func TestParseMemoryString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"megabytes", "512M", 512, false},
		{"megabytes lowercase", "512m", 512, false},
		{"gigabytes", "2G", 2048, false},
		{"gigabytes lowercase", "2g", 2048, false},
		{"kilobytes", "2048K", 2, false},
		{"kilobytes lowercase", "2048k", 2, false},
		{"plain number", "128", 128, false},
		{"with spaces", " 256M ", 256, false},
		{"empty", "", 0, true},
		{"invalid", "abcM", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoryString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseMemoryString(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSystemdAvailable(t *testing.T) {
	// Just verify it doesn't panic
	_ = isSystemdAvailable()
}
