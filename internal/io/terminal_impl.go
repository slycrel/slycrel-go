package io

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/term"
)

// LocalTerminal implements IOProvider for a local terminal session using raw ANSI output.
type LocalTerminal struct {
	reader    *bufio.Reader
	dataDir   string // path to data/ directory for ANSI art files
	connected bool
	oldState  *term.State
}

// NewLocalTerminal creates a new local terminal I/O provider.
// dataDir is the path to the data/ directory containing ANSI art files.
func NewLocalTerminal(dataDir string) *LocalTerminal {
	return &LocalTerminal{
		reader:    bufio.NewReader(os.Stdin),
		dataDir:   dataDir,
		connected: true,
	}
}

// EnableRawMode puts the terminal into raw mode for single-keypress input.
// Call RestoreTerminal() to undo this.
func (t *LocalTerminal) EnableRawMode() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	t.oldState = oldState
	return nil
}

// RestoreTerminal restores the terminal to its original state.
func (t *LocalTerminal) RestoreTerminal() {
	if t.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), t.oldState)
		t.oldState = nil
	}
}

func (t *LocalTerminal) Outln(text string, newline bool, color int) {
	width := t.termWidth()
	if width <= 0 {
		width = 80
	}

	// Split on embedded newlines first
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Word-wrap each line
		wrapped := wordWrap(line, width-1) // -1 to avoid terminal auto-wrap
		for j, wl := range wrapped {
			fmt.Print(ColorCode(color) + wl + Reset)
			// Add \r\n between wrapped segments and between embedded newlines
			if j < len(wrapped)-1 {
				fmt.Print("\r\n")
			}
		}
		if i < len(lines)-1 {
			fmt.Print("\r\n")
		}
	}
	if newline {
		fmt.Print("\r\n")
	}
}

// termWidth returns the current terminal width.
func (t *LocalTerminal) termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return w
}

// wordWrap breaks a line into multiple lines at word boundaries.
func wordWrap(text string, width int) []string {
	if width <= 0 || len(text) <= width {
		return []string{text}
	}

	var lines []string
	for len(text) > width {
		// Find the last space within width
		breakAt := strings.LastIndex(text[:width], " ")
		if breakAt <= 0 {
			// No space found - break at width
			breakAt = width
		}
		lines = append(lines, text[:breakAt])
		text = strings.TrimLeft(text[breakAt:], " ")
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}

func (t *LocalTerminal) ANSICode(code string) {
	fmt.Print(ESC + code)
}

func (t *LocalTerminal) Cr() {
	fmt.Print("\r\n")
}

func (t *LocalTerminal) ClearScreen() {
	fmt.Print(ClearScreenCode())
}

func (t *LocalTerminal) LettersPrompt(prompt string, allowed string, maxChars int, capitalize bool, showInput bool) string {
	fmt.Print(ColorCode(1) + prompt + Reset + " ")

	// Single character prompt: use instant keypress (no Enter needed)
	if maxChars == 1 {
		for {
			ch := t.readOneKey()
			if ch == 0 {
				return ""
			}
			input := strings.ToUpper(string(ch))
			if capitalize {
				// already uppercase
			}

			if allowed == "" {
				if showInput {
					fmt.Print(input)
				}
				fmt.Print("\r\n")
				return input
			}

			if strings.Contains(strings.ToUpper(allowed), input) {
				if showInput {
					fmt.Print(input)
				}
				fmt.Print("\r\n")
				return input
			}
		}
	}

	// Multi-character: read full line
	for {
		line := t.readLineRaw(maxChars)
		if len(line) == 0 {
			continue
		}

		input := line
		if capitalize {
			input = strings.ToUpper(input)
		}

		if allowed == "" {
			return input
		}

		ch := string(input[0])
		if strings.Contains(strings.ToUpper(allowed), strings.ToUpper(ch)) {
			return ch
		}
	}
}

func (t *LocalTerminal) NumbersPrompt(prompt string, min int, max int) int {
	for {
		fmt.Print(ColorCode(1) + prompt + Reset + " ")
		line := t.readLineRaw(10)
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < min || n > max {
			fmt.Print(Colorize(fmt.Sprintf("Please enter a number between %d and %d.", min, max), 6) + "\r\n")
			continue
		}
		return n
	}
}

func (t *LocalTerminal) YesNoQuestion(prompt string) bool {
	fmt.Print(ColorCode(1) + prompt + Reset + " ")
	ch := t.readOneKey()
	input := strings.ToUpper(string(ch))
	fmt.Print(input + "\r\n")
	return input != "N"
}

func (t *LocalTerminal) PausePrompt(prompt string) {
	fmt.Print(ColorCode(1) + prompt + Reset)
	t.readOneKey()
	fmt.Print("\r\n")
}

func (t *LocalTerminal) ReadLine(prompt string, maxLen int) string {
	fmt.Print(ColorCode(1) + prompt + Reset + " ")
	return t.readLineRaw(maxLen)
}

func (t *LocalTerminal) ShowANSIFile(name string) error {
	path := filepath.Join(t.dataDir, "ansi", name+".ans")
	data, err := os.ReadFile(path)
	if err != nil {
		// Not an error if the file doesn't exist - just skip it
		return nil
	}
	content := string(data)
	// Convert \e[ placeholder to real ESC[ sequences.
	// This lets .ans files be edited as plain text without embedding raw escape bytes.
	content = strings.ReplaceAll(content, `\e[`, "\033[")
	// Convert any bare \n to \r\n for raw mode
	content = strings.ReplaceAll(content, "\n", "\r\n")
	fmt.Print(content)
	return nil
}

func (t *LocalTerminal) IsConnected() bool {
	return t.connected
}

// Disconnect marks the session as disconnected.
func (t *LocalTerminal) Disconnect() {
	t.connected = false
}

// readLineRaw reads a line of input, handling backspace in raw mode.
func (t *LocalTerminal) readLineRaw(maxLen int) string {
	var buf []byte
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			t.connected = false
			return string(buf)
		}

		ch := b[0]

		// Enter
		if ch == '\r' || ch == '\n' {
			fmt.Print("\r\n")
			return string(buf)
		}

		// Ctrl-C / Ctrl-D
		if ch == 3 || ch == 4 {
			t.connected = false
			return ""
		}

		// Backspace / Delete
		if ch == 127 || ch == 8 {
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}
			continue
		}

		// Escape sequences (arrow keys etc.) - consume and ignore
		if ch == 27 {
			// Read the next bytes of the escape sequence
			next := make([]byte, 1)
			os.Stdin.Read(next)
			if next[0] == '[' {
				// CSI sequence - read until we get a letter
				for {
					seq := make([]byte, 1)
					os.Stdin.Read(seq)
					if unicode.IsLetter(rune(seq[0])) {
						break
					}
				}
			}
			continue
		}

		// Only accept printable characters
		if ch >= 32 && ch < 127 {
			if maxLen > 0 && len(buf) >= maxLen {
				continue
			}
			buf = append(buf, ch)
			fmt.Print(string(ch))
		}
	}
}

// readOneKey reads a single keypress in raw mode.
func (t *LocalTerminal) readOneKey() byte {
	b := make([]byte, 1)
	_, err := os.Stdin.Read(b)
	if err != nil {
		t.connected = false
		return 0
	}
	if b[0] == 3 || b[0] == 4 {
		t.connected = false
	}
	return b[0]
}
