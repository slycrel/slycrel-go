package io

// Protocol version — bump when message schema changes. Server rejects mismatched clients.
const WSProtocolVersion = 1

// MsgType identifies the kind of WebSocket message.
type MsgType string

const (
	// Server → Client
	MsgTypeOutput  MsgType = "output"   // Outln / Cr / ANSICode / ClearScreen
	MsgTypePrompt  MsgType = "prompt"   // LettersPrompt / NumbersPrompt / YesNoQuestion / PausePrompt / ReadLine
	MsgTypeANSIArt MsgType = "ansi_art" // ShowANSIFile — raw ANSI bytes

	// Client → Server
	MsgTypeInput MsgType = "input" // response to any prompt

	// Bidirectional / handshake
	MsgTypeHello MsgType = "hello" // client sends on connect; server echoes with version check
	MsgTypeError MsgType = "error" // server sends on protocol violation
)

// PromptKind distinguishes input modalities so the browser can render the right widget.
type PromptKind string

const (
	PromptKindLetters  PromptKind = "letters"   // constrained single/multi char
	PromptKindNumber   PromptKind = "number"     // integer in [min, max]
	PromptKindYesNo    PromptKind = "yesno"      // y/n boolean
	PromptKindPause    PromptKind = "pause"      // any key to continue
	PromptKindReadLine PromptKind = "readline"   // free-form text line
)

// ServerMsg is every message the server sends to the client.
type ServerMsg struct {
	Version int     `json:"v"`              // WSProtocolVersion
	Type    MsgType `json:"type"`
	Seq     int     `json:"seq"`            // monotonic; client echoes this in InputMsg

	// output fields (MsgTypeOutput)
	Text    string `json:"text,omitempty"`
	Newline bool   `json:"newline,omitempty"`
	Color   int    `json:"color,omitempty"` // 1-6 matching IOProvider color index
	ANSI    string `json:"ansi,omitempty"`  // raw ANSI code (ANSICode calls)

	// prompt fields (MsgTypePrompt)
	Prompt     string     `json:"prompt,omitempty"`
	Kind       PromptKind `json:"kind,omitempty"`
	Allowed    string     `json:"allowed,omitempty"`    // LettersPrompt
	MaxChars   int        `json:"max_chars,omitempty"`  // LettersPrompt
	Capitalize bool       `json:"capitalize,omitempty"` // LettersPrompt
	ShowInput  bool       `json:"show_input,omitempty"` // LettersPrompt
	Min        int        `json:"min,omitempty"`        // NumbersPrompt
	Max        int        `json:"max,omitempty"`        // NumbersPrompt
	MaxLen     int        `json:"max_len,omitempty"`    // ReadLine

	// ansi_art fields (MsgTypeANSIArt)
	ANSIArtName string `json:"ansi_art_name,omitempty"`
	ANSIArtData string `json:"ansi_art_data,omitempty"` // base64-encoded ANSI file contents

	// error fields (MsgTypeError)
	ErrMsg string `json:"err,omitempty"`
}

// ClientMsg is every message the client sends to the server.
type ClientMsg struct {
	Version int     `json:"v"`    // WSProtocolVersion — server validates
	Type    MsgType `json:"type"`
	Seq     int     `json:"seq"`  // echo of the prompt Seq this answers

	// hello fields (MsgTypeHello)
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// input fields (MsgTypeInput)
	Value string `json:"value,omitempty"` // raw string; server parses per PromptKind
}
