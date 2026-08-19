// Package hwclock manages hardware clock (RTC).
// Equivalent to ansible.posix.hwclock module.
package hwclock

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Action  string `json:"action,omitempty"`
	Time    string `json:"time,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GetResult is returned by Get.
type GetResult struct {
	Status string `json:"status"`
	HWTime string `json:"hw_time"`
	SWTime string `json:"sw_time"`
	RTC    string `json:"rtc,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Get returns the current hardware and software clock times.
func Get() GetResult {
	cmd := exec.Command("hwclock", "--show")
	out, err := cmd.Output()
	if err != nil {
		return GetResult{Status: "failed", Error: fmt.Sprintf("hwclock: %v", err)}
	}

	hwTime := strings.TrimSpace(string(out))
	swTime := time.Now().Format("2006-01-02 15:04:05")

	return GetResult{
		Status: "success",
		HWTime: hwTime,
		SWTime: swTime,
	}
}

// Set sets the hardware clock to the current system time.
func Set() Result {
	cmd := exec.Command("hwclock", "--systohc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("hwclock: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Action: "systohc", Time: time.Now().Format("2006-01-02 15:04:05")}
}

// HCToSys sets the system clock from the hardware clock.
func HCToSys() Result {
	cmd := exec.Command("hwclock", "--hctosys")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("hwclock: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Action: "hctosys", Time: time.Now().Format("2006-01-02 15:04:05")}
}

// SetTime sets the hardware clock to a specific time.
func SetTime(timeStr string) Result {
	if timeStr == "" {
		return Result{Status: "failed", Error: "time is required"}
	}

	cmd := exec.Command("hwclock", "--set", "--date="+timeStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("hwclock: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Action: "set", Time: timeStr}
}
