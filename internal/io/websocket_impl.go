package io

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

const (
	wsReadTimeout  = 5 * time.Minute // idle client disconnect detection
	wsWriteTimeout = 30 * time.Second
)

// WSSession implements IOProvider over a WebSocket connection.
//
// The game state machine drives a strict request-response loop:
//   server sends prompt → client sends exactly one input → server advances state
//
// Concurrency model: one goroutine owns the session (game loop). Send helpers
// are safe to call from that goroutine only. No concurrent reads are issued.
type WSSession struct {
	conn    *websocket.Conn
	dataDir string

	mu        sync.Mutex // guards connected
	connected bool

	seq atomic.Int64 // monotonic message counter
}

// NewWSSession wraps an established WebSocket connection as an IOProvider.
// dataDir is the path to the data/ directory (for ANSI art files).
func NewWSSession(conn *websocket.Conn, dataDir string) *WSSession {
	return &WSSession{
		conn:      conn,
		dataDir:   dataDir,
		connected: true,
	}
}

// IsConnected returns true while the WebSocket connection is alive.
func (s *WSSession) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

func (s *WSSession) disconnect() {
	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()
}

// nextSeq returns a monotonically increasing sequence number.
func (s *WSSession) nextSeq() int {
	return int(s.seq.Add(1))
}

// send marshals and writes a ServerMsg. On write error, marks session disconnected.
func (s *WSSession) send(msg ServerMsg) error {
	msg.Version = WSProtocolVersion
	if msg.Seq == 0 {
		msg.Seq = s.nextSeq()
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.conn.SetDeadline(time.Now().Add(wsWriteTimeout))
	_, err = s.conn.Write(data)
	if err != nil {
		s.disconnect()
		return err
	}
	return nil
}

// recv reads one ClientMsg. On read error, marks session disconnected.
func (s *WSSession) recv() (ClientMsg, error) {
	s.conn.SetDeadline(time.Now().Add(wsReadTimeout))
	var raw []byte
	if _, err := s.conn.Read(raw); err != nil {
		s.disconnect()
		return ClientMsg{}, err
	}
	var msg ClientMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return ClientMsg{}, fmt.Errorf("protocol error: %w", err)
	}
	if msg.Version != WSProtocolVersion {
		_ = s.send(ServerMsg{
			Type:   MsgTypeError,
			ErrMsg: fmt.Sprintf("protocol version mismatch: server=%d client=%d", WSProtocolVersion, msg.Version),
		})
		s.disconnect()
		return ClientMsg{}, fmt.Errorf("version mismatch")
	}
	return msg, nil
}

// sendPromptAndRecv sends a prompt ServerMsg and waits for the matching InputMsg.
func (s *WSSession) sendPromptAndRecv(msg ServerMsg) (string, error) {
	seq := s.nextSeq()
	msg.Seq = seq
	if err := s.send(msg); err != nil {
		return "", err
	}
	for {
		reply, err := s.recv()
		if err != nil {
			return "", err
		}
		if reply.Type == MsgTypeInput && reply.Seq == seq {
			return reply.Value, nil
		}
		// Ignore out-of-order or unexpected messages — strict ordering enforced.
	}
}

// --- IOProvider implementation ---

// Outln sends a text output message to the client.
func (s *WSSession) Outln(text string, newline bool, color int) {
	if !s.IsConnected() {
		return
	}
	_ = s.send(ServerMsg{
		Type:    MsgTypeOutput,
		Text:    text,
		Newline: newline,
		Color:   color,
	})
}

// ANSICode sends a raw ANSI escape sequence (without ESC[ prefix).
func (s *WSSession) ANSICode(code string) {
	if !s.IsConnected() {
		return
	}
	_ = s.send(ServerMsg{
		Type: MsgTypeOutput,
		ANSI: code,
	})
}

// Cr sends a newline.
func (s *WSSession) Cr() {
	s.Outln("", true, 0)
}

// ClearScreen signals the client to clear its display.
func (s *WSSession) ClearScreen() {
	if !s.IsConnected() {
		return
	}
	_ = s.send(ServerMsg{
		Type: MsgTypeOutput,
		ANSI: "2J", // matches ClearScreen ANSI sequence
	})
}

// LettersPrompt sends a constrained-input prompt and returns the chosen character(s).
func (s *WSSession) LettersPrompt(prompt string, allowed string, maxChars int, capitalize bool, showInput bool) string {
	if !s.IsConnected() {
		return ""
	}
	val, err := s.sendPromptAndRecv(ServerMsg{
		Type:       MsgTypePrompt,
		Prompt:     prompt,
		Kind:       PromptKindLetters,
		Allowed:    allowed,
		MaxChars:   maxChars,
		Capitalize: capitalize,
		ShowInput:  showInput,
	})
	if err != nil {
		return ""
	}
	if capitalize {
		val = strings.ToUpper(val)
	}
	if len(val) > maxChars && maxChars > 0 {
		val = val[:maxChars]
	}
	return val
}

// NumbersPrompt sends a numeric prompt and returns the validated integer.
func (s *WSSession) NumbersPrompt(prompt string, min int, max int) int {
	if !s.IsConnected() {
		return min
	}
	for {
		val, err := s.sendPromptAndRecv(ServerMsg{
			Type:   MsgTypePrompt,
			Prompt: prompt,
			Kind:   PromptKindNumber,
			Min:    min,
			Max:    max,
		})
		if err != nil {
			return min
		}
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil || n < min || n > max {
			_ = s.send(ServerMsg{
				Type:   MsgTypeError,
				ErrMsg: fmt.Sprintf("enter a number between %d and %d", min, max),
			})
			continue
		}
		return n
	}
}

// YesNoQuestion sends a yes/no prompt and returns true for yes.
func (s *WSSession) YesNoQuestion(prompt string) bool {
	if !s.IsConnected() {
		return false
	}
	for {
		val, err := s.sendPromptAndRecv(ServerMsg{
			Type:   MsgTypePrompt,
			Prompt: prompt,
			Kind:   PromptKindYesNo,
		})
		if err != nil {
			return false
		}
		v := strings.ToLower(strings.TrimSpace(val))
		if v == "y" || v == "yes" {
			return true
		}
		if v == "n" || v == "no" {
			return false
		}
	}
}

// PausePrompt sends a pause prompt and waits for any key from the client.
func (s *WSSession) PausePrompt(prompt string) {
	if !s.IsConnected() {
		return
	}
	_, _ = s.sendPromptAndRecv(ServerMsg{
		Type:   MsgTypePrompt,
		Prompt: prompt,
		Kind:   PromptKindPause,
	})
}

// ReadLine sends a free-form text prompt and returns the entered line.
func (s *WSSession) ReadLine(prompt string, maxLen int) string {
	if !s.IsConnected() {
		return ""
	}
	val, err := s.sendPromptAndRecv(ServerMsg{
		Type:   MsgTypePrompt,
		Prompt: prompt,
		Kind:   PromptKindReadLine,
		MaxLen: maxLen,
	})
	if err != nil {
		return ""
	}
	if maxLen > 0 && len(val) > maxLen {
		val = val[:maxLen]
	}
	return strings.TrimSpace(val)
}

// ShowANSIFile reads an ANSI art file from dataDir and sends its contents to the client.
func (s *WSSession) ShowANSIFile(name string) error {
	if !s.IsConnected() {
		return nil
	}
	path := filepath.Join(s.dataDir, "ansi", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ansi file %q: %w", name, err)
	}
	return s.send(ServerMsg{
		Type:        MsgTypeANSIArt,
		ANSIArtName: name,
		ANSIArtData: base64.StdEncoding.EncodeToString(data),
	})
}

// --- WebSocket upgrade helper (for use in cmd/server) ---

// WSUpgrader returns an http.Handler that upgrades HTTP connections to WebSocket
// and invokes handler with the established WSSession.
//
// Authentication is the caller's responsibility: inspect the *http.Request
// before calling this, or validate credentials in the first ClientMsg (MsgTypeHello).
func WSUpgrader(dataDir string, handler func(sess *WSSession, username string)) http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		// Expect MsgTypeHello as the first message for auth + version check.
		conn.SetDeadline(time.Now().Add(30 * time.Second))
		var raw []byte
		if _, err := conn.Read(raw); err != nil {
			return
		}
		var hello ClientMsg
		if err := json.Unmarshal(raw, &hello); err != nil || hello.Type != MsgTypeHello {
			return
		}
		if hello.Version != WSProtocolVersion {
			data, _ := json.Marshal(ServerMsg{
				Version: WSProtocolVersion,
				Type:    MsgTypeError,
				ErrMsg:  fmt.Sprintf("protocol version mismatch: server=%d client=%d", WSProtocolVersion, hello.Version),
			})
			conn.Write(data) //nolint:errcheck
			return
		}

		sess := NewWSSession(conn, dataDir)
		handler(sess, hello.Username)
	})
}
