package io

import "fmt"

// ANSI escape code helpers.
// The original game used ANSI extensively for colors and cursor positioning.

const (
	ESC   = "\033["
	Reset = ESC + "0m"
)

// Color codes matching the original game's color index system.
// G_Outln used a color parameter (1-6) mapped to ANSI codes.
var colorCodes = map[int]string{
	0: ESC + "0m",       // reset/default
	1: ESC + "0;36m",    // cyan (info/prompts)
	2: ESC + "1;37m",    // bright white
	3: ESC + "1;32m",    // bright green (success)
	4: ESC + "1;33m",    // bright yellow (character names)
	5: ESC + "1;35m",    // bright magenta
	6: ESC + "1;31m",    // bright red (errors/warnings)
}

// ColorCode returns the ANSI escape sequence for the given color index.
func ColorCode(index int) string {
	if code, ok := colorCodes[index]; ok {
		return code
	}
	return colorCodes[0]
}

// Colorize wraps text in the given color and resets after.
func Colorize(text string, colorIndex int) string {
	return ColorCode(colorIndex) + text + Reset
}

// Bold wraps text in bold.
func Bold(text string) string {
	return ESC + "1m" + text + Reset
}

// CursorTo moves the cursor to row, col (1-based).
func CursorTo(row, col int) string {
	return fmt.Sprintf("%s%d;%dH", ESC, row, col)
}

// ClearScreenCode returns the ANSI clear screen sequence.
func ClearScreenCode() string {
	return ESC + "2J" + ESC + "1;1H"
}
