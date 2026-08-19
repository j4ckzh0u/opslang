package sysfs

import (
	"encoding/json"
	"testing"
)

func TestRead_EmptyPath(t *testing.T) {
	_, err := Read("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRead_InvalidPath(t *testing.T) {
	_, err := Read("/etc/passwd")
	if err == nil {
		t.Fatal("expected error for path not under /sys")
	}
}

func TestWrite_EmptyPath(t *testing.T) {
	_, err := Write("", "value")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestWrite_InvalidPath(t *testing.T) {
	_, err := Write("/etc/test", "value")
	if err == nil {
		t.Fatal("expected error for path not under /sys")
	}
}

func TestExists_EmptyPath(t *testing.T) {
	_, err := Exists("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestGet_EmptyPath(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestList_EmptyPath(t *testing.T) {
	_, err := List("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSetDevicePower_EmptyArgs(t *testing.T) {
	_, err := SetDevicePower("", "auto")
	if err == nil {
		t.Fatal("expected error for empty device path")
	}
	_, err = SetDevicePower("/sys/devices/test", "")
	if err == nil {
		t.Fatal("expected error for empty state")
	}
}

func TestGetDevicePower_EmptyPath(t *testing.T) {
	_, err := GetDevicePower("")
	if err == nil {
		t.Fatal("expected error for empty device path")
	}
}

func TestSetKernelParameter_EmptyParam(t *testing.T) {
	_, err := SetKernelParameter("", "1")
	if err == nil {
		t.Fatal("expected error for empty parameter")
	}
}

func TestGetKernelParameter_EmptyParam(t *testing.T) {
	_, err := GetKernelParameter("")
	if err == nil {
		t.Fatal("expected error for empty parameter")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Path: "/sys/class/net/eth0/mtu", Changed: true, Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestAttributeInfoJSON(t *testing.T) {
	attr := AttributeInfo{
		Path:  "/sys/class/net/eth0/mtu",
		Value: "1500",
		Mode:  "-r--r--r--",
	}
	b, err := json.Marshal(attr)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Attributes: []AttributeInfo{
			{Path: "/sys/class/net/eth0/mtu", Value: "1500"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestReadSysfs(t *testing.T) {
	// Try to read a real sysfs attribute that should exist on most systems
	value, err := Read("/sys/class/net/lo/mtu")
	if err != nil {
		t.Logf("Read /sys/class/net/lo/mtu failed (may not exist on this system): %v", err)
	} else {
		t.Logf("Read /sys/class/net/lo/mtu: %s", value)
	}
}
