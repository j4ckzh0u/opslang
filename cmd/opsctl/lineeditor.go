// lineeditor implements the interactive line reader for the REPL: raw-mode
// terminal input with command history (Up/Down), cursor movement
// (Left/Right/Home/End), and the usual editing keys. Non-TTY stdin (pipes,
// CI, tests) falls back to plain buffered reads so scripted use of the REPL
// keeps working byte-for-byte.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// errInterrupted is returned when the user presses Ctrl+C: the current line
// is discarded but the REPL itself keeps running.
var errInterrupted = errors.New("interrupted")

// history is the per-session command history shared by every prompt of one
// REPL run. Empty submissions and exact duplicates of the newest entry are
// not stored (same convention as shell HISTCONTROL=erasedups-lite).
type history struct {
	entries []string
	// navIdx is the cursor into entries while browsing; -1 means the live
	// (not yet submitted) line.
	navIdx int
	// live holds the line being typed so Down can restore it after browsing.
	live string
}

func (h *history) add(line string) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return
	}
	if n := len(h.entries); n > 0 && h.entries[n-1] == line {
		return
	}
	h.entries = append(h.entries, line)
	// Bound memory the same way shells do; 500 entries is plenty for a
	// session and keeps the slice from growing without limit.
	if len(h.entries) > 500 {
		h.entries = h.entries[len(h.entries)-500:]
	}
	h.navIdx = -1
}

// lineReader reads one logical input line. On a TTY it renders the prompt
// and handles editing; otherwise it degrades to a buffered read that still
// honors the prompt (for observability) but expects pre-formed lines.
type lineReader struct {
	prompt   string
	hist     *history
	fallback *bufio.Reader
	rawState *term.State
	isTTY    bool
}

func newLineReader(prompt string, hist *history) *lineReader {
	return &lineReader{prompt: prompt, hist: hist, isTTY: term.IsTerminal(int(os.Stdin.Fd()))}
}

// init puts the terminal into raw mode for the whole REPL session.
// Entering raw mode per-line left a restore/re-enter window between lines
// in which the cooked line discipline consumed bytes that arrived in it —
// a whole "print(" prefix could vanish. One session-long mode has no gap.
// Raw mode also stops the terminal from translating "\n" to CRLF, so all
// REPL output goes through emit() which adds the "\r".
func (r *lineReader) init() error {
	if !r.isTTY {
		return nil
	}
	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	r.rawState = state
	return nil
}

func (r *lineReader) close() {
	if r.rawState != nil {
		term.Restore(int(os.Stdin.Fd()), r.rawState)
		r.rawState = nil
	}
}

// emit prints with newline translation for raw mode: ONLCR is off, so a
// bare "\n" would move down without returning to column 0 and every
// following line would render as a staircase.
func (r *lineReader) emit(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	if r.isTTY {
		text = strings.ReplaceAll(text, "\n", "\r\n")
	}
	fmt.Print(text)
}

// read returns one submitted line (newline-stripped). io.EOF means the
// input stream is exhausted (Ctrl+D on an empty line); errInterrupted means
// the user abandoned the line and wants a fresh prompt.
func (r *lineReader) read() (string, error) {
	if !r.isTTY {
		if r.fallback == nil {
			r.fallback = bufio.NewReader(os.Stdin)
		}
		fmt.Print(r.prompt)
		line, err := r.fallback.ReadString('\n')
		if err != nil {
			if err == io.EOF && line != "" {
				return line, nil // last line without trailing newline
			}
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	buf := []rune{}
	cursor := 0
	r.hist.navIdx = -1
	r.hist.live = ""
	r.draw(buf, cursor)

	for {
		var key [16]byte
		n, err := os.Stdin.Read(key[:])
		if err != nil {
			if err == io.EOF {
				if len(buf) == 0 {
					return "", io.EOF
				}
				// Ctrl+D mid-line submits the line (readline behavior).
				r.emit("\n")
				return string(buf), nil
			}
			return "", err
		}
		runes := []rune(string(key[:n]))

		for i := 0; i < len(runes); i++ {
			ru := runes[i]
			switch ru {
			case '\r', '\n':
				r.emit("\n")
				return string(buf), nil

			case 0x03: // Ctrl+C
				r.emit("^C\n")
				return "", errInterrupted

			case 0x04: // Ctrl+D
				if len(buf) == 0 {
					r.emit("\n")
					return "", io.EOF
				}
				buf = append(buf[:cursor], buf[cursor+1:]...) // no-op delete at EOL; readline beeps, we just redraw
				r.draw(buf, cursor)

			case 0x7f, 0x08: // Backspace / Ctrl+H
				if cursor > 0 {
					buf = append(buf[:cursor-1], buf[cursor:]...)
					cursor--
				}
				r.draw(buf, cursor)

			case 0x01: // Ctrl+A -> home
				cursor = 0
				r.draw(buf, cursor)

			case 0x05: // Ctrl+E -> end
				cursor = len(buf)
				r.draw(buf, cursor)

			case 0x15: // Ctrl+U -> delete to start of line
				buf = append([]rune{}, buf[cursor:]...)
				cursor = 0
				r.draw(buf, cursor)

			case 0x0b: // Ctrl+K -> delete to end of line
				buf = buf[:cursor]
				r.draw(buf, cursor)

			case 0x0c: // Ctrl+L -> clear screen, redraw
				r.emit("\x1b[2J\x1b[H")
				r.draw(buf, cursor)

			case '\t': // Tab: insert spaces (no completion yet)
				buf = append(buf[:cursor], append([]rune("    "), buf[cursor:]...)...)
				cursor += 4
				r.draw(buf, cursor)

			case 0x1b: // ESC sequence
				// Reads from a terminal are allowed to split an escape sequence
				// across syscalls, especially over SSH. Complete the sequence
				// before decoding it instead of dropping a lone ESC byte.
				for len(runes)-i < 3 {
					var tail [8]byte
					next, readErr := os.Stdin.Read(tail[:])
					if next > 0 {
						runes = append(runes, []rune(string(tail[:next]))...)
					}
					if readErr != nil || next == 0 {
						break
					}
				}
				if i+2 < len(runes) && (runes[i+1] == '[' || runes[i+1] == 'O') {
					final := runes[i+2]
					i += 2
					switch final {
					case 'A': // Up
						buf, cursor = r.histPrev(buf)
					case 'B': // Down
						buf, cursor = r.histNext(buf)
					case 'C': // Right
						if cursor < len(buf) {
							cursor++
						}
					case 'D': // Left
						if cursor > 0 {
							cursor--
						}
					case 'H': // Home
						cursor = 0
					case 'F': // End
						cursor = len(buf)
					case '3': // Delete (ESC [ 3 ~)
						if i+3 < len(runes) && runes[i+3] == '~' {
							i++
						}
						if cursor < len(buf) {
							buf = append(buf[:cursor], buf[cursor+1:]...)
						}
					case '1': // Home variant (ESC [ 1 ~)
						if i+3 < len(runes) && runes[i+3] == '~' {
							i++
						}
						cursor = 0
					case '4': // End variant (ESC [ 4 ~)
						if i+3 < len(runes) && runes[i+3] == '~' {
							i++
						}
						cursor = len(buf)
					}
					r.draw(buf, cursor)
				} else {
					// Lone ESC: ignore (meta-key prefix without a
					// recognized sequence).
				}

			default:
				if ru >= 0x20 {
					buf = append(buf[:cursor], append([]rune{ru}, buf[cursor:]...)...)
					cursor++
					r.draw(buf, cursor)
				}
				// Other control bytes are ignored rather than inserted.
			}
		}
	}
}

// histPrev loads the previous entry (Up). The live line is stashed so Down
// can bring it back.
func (r *lineReader) histPrev(buf []rune) ([]rune, int) {
	h := r.hist
	if len(h.entries) == 0 {
		return buf, len(buf)
	}
	if h.navIdx == -1 {
		h.live = string(buf)
		h.navIdx = len(h.entries) - 1
	} else if h.navIdx > 0 {
		h.navIdx--
	}
	line := []rune(h.entries[h.navIdx])
	return line, len(line)
}

// histNext loads the next entry (Down); past the newest entry it restores
// the stashed live line.
func (r *lineReader) histNext(buf []rune) ([]rune, int) {
	h := r.hist
	if h.navIdx == -1 {
		return buf, len(buf)
	}
	h.navIdx++
	if h.navIdx >= len(h.entries) {
		h.navIdx = -1
		line := []rune(h.live)
		return line, len(line)
	}
	line := []rune(h.entries[h.navIdx])
	return line, len(line)
}

// draw repaints prompt + buffer and parks the cursor. Rendering is
// redraw-from-scratch: \r, clear to end of line, then reprint and move the
// cursor left from the end by the number of runes after the cursor. This
// stays correct when the edit shrinks the line and needs no scrollback
// math. Prompt is assumed to be single-line and ANSI-clean.
func (r *lineReader) draw(buf []rune, cursor int) {
	after := len(buf) - cursor
	fmt.Printf("\r%s%s\x1b[0K", r.prompt, string(buf))
	if after > 0 {
		fmt.Printf("\x1b[%dD", after)
	}
}
