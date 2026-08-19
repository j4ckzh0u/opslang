// Package pause provides execution pause functionality.
// Equivalent to ansible.builtin.pause module.
package pause

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Result is returned by all functions.
type Result struct {
	Status    string `json:"status"`
	DurationMs int64 `json:"duration_ms"`
	Message   string `json:"message,omitempty"`
	Input     string `json:"input,omitempty"`
	Stopped   bool   `json:"stopped"`
	Error     string `json:"error,omitempty"`
}

// Seconds pauses execution for a specified number of seconds.
func Seconds(duration int) Result {
	if duration <= 0 {
		return Result{Status: "failed", Error: "duration must be positive"}
	}

	start := time.Now()
	time.Sleep(time.Duration(duration) * time.Second)
	return Result{
		Status:    "success",
		DurationMs: time.Since(start).Milliseconds(),
		Message:   fmt.Sprintf("paused for %d seconds", duration),
	}
}

// Prompt pauses execution and waits for user input.
func Prompt(message string) Result {
	if message == "" {
		message = "Press Enter to continue"
	}

	fmt.Print(message + ": ")
	start := time.Now()

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		return Result{
			Status:    "success",
			DurationMs: time.Since(start).Milliseconds(),
			Message:   message,
			Input:     input,
		}
	}

	if err := scanner.Err(); err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("read input: %v", err)}
	}

	return Result{Status: "failed", Error: "no input received"}
}

// PromptWithDefault pauses and waits for input with a default value.
func PromptWithDefault(message, defaultVal string) Result {
	if message == "" {
		message = "Press Enter to continue"
	}

	fmt.Printf("%s [%s]: ", message, defaultVal)
	start := time.Now()

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			input = defaultVal
		}
		return Result{
			Status:    "success",
			DurationMs: time.Since(start).Milliseconds(),
			Message:   message,
			Input:     input,
		}
	}

	return Result{Status: "failed", Error: "no input received"}
}
