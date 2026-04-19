package io

// IOProvider abstracts all user I/O, replacing the Hermes BBS framework callbacks.
// Implementations: local terminal (now), WebSocket/telnet (future).
type IOProvider interface {
	// Outln outputs text. color maps to the original's color index:
	// 1=cyan, 2=white, 3=green, 4=yellow, 5=magenta, 6=red
	Outln(text string, newline bool, color int)

	// ANSICode sends a raw ANSI escape sequence (without the ESC[ prefix).
	ANSICode(code string)

	// Cr outputs a carriage return / newline.
	Cr()

	// ClearScreen clears the terminal.
	ClearScreen()

	// LettersPrompt displays a prompt and waits for the user to type one of the
	// allowed characters. Returns the uppercase character chosen.
	// maxChars is the maximum input length.
	// capitalize: if true, auto-uppercase input.
	// showInput: if true, echo the typed character.
	LettersPrompt(prompt string, allowed string, maxChars int, capitalize bool, showInput bool) string

	// NumbersPrompt displays a prompt and reads a number within range.
	NumbersPrompt(prompt string, min int, max int) int

	// YesNoQuestion displays a yes/no prompt and returns true for yes.
	YesNoQuestion(prompt string) bool

	// PausePrompt displays a message and waits for any key.
	PausePrompt(prompt string)

	// ReadLine reads a full line of text input (for names, etc.).
	ReadLine(prompt string, maxLen int) string

	// ShowANSIFile displays an ANSI art file by name from the data/ansi directory.
	ShowANSIFile(name string) error

	// RenderScene draws a structured grid scene (terrain tiles + entity overlay + HUD).
	// Terminal implementations render via ANSI; tilemap-aware clients render via their
	// engine's native facilities.
	RenderScene(scene Scene)

	// IsConnected returns true if the user session is still active.
	IsConnected() bool
}
