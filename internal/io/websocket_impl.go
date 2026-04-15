package io

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage is the JSON envelope for all WebSocket messages.
type WSMessage struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Prompt  string `json:"prompt,omitempty"`
	Allowed string `json:"allowed,omitempty"`
	Min     int    `json:"min,omitempty"`
	Max     int    `json:"max,omitempty"`
	MaxLen  int    `json:"maxLen,omitempty"`
	Color   int    `json:"color,omitempty"`
}

// WSClientMessage is a message from the browser back to the server.
type WSClientMessage struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// WebSocketTerminal implements IOProvider over a WebSocket connection.
// The game engine calls the same IOProvider methods as the terminal version;
// this implementation serialises them to JSON and exchanges them with the
// browser client over the WebSocket.
type WebSocketTerminal struct {
	conn      *websocket.Conn
	mu        sync.Mutex   // protects conn writes
	replyCh   chan string   // responses from the client
	connected bool
	dataDir   string
}

// NewWebSocketTerminal creates a WebSocket-backed IOProvider.
// dataDir is the path to the data/ directory (for ANSI art files).
func NewWebSocketTerminal(conn *websocket.Conn, dataDir string) *WebSocketTerminal {
	wst := &WebSocketTerminal{
		conn:      conn,
		replyCh:   make(chan string, 1),
		connected: true,
		dataDir:   dataDir,
	}
	go wst.readLoop()
	return wst
}

// readLoop pumps messages from the browser into replyCh.
func (w *WebSocketTerminal) readLoop() {
	defer func() {
		w.mu.Lock()
		w.connected = false
		w.mu.Unlock()
		// Drain/unblock any waiting prompt.
		select {
		case w.replyCh <- "":
		default:
		}
	}()

	for {
		_, data, err := w.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg WSClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("ws: bad client message: %v", err)
			continue
		}
		if msg.Type == "input" {
			select {
			case w.replyCh <- msg.Value:
			default:
				// Drop if nobody is waiting.
			}
		}
	}
}

// send serialises and sends a message to the client.
func (w *WebSocketTerminal) send(msg WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

// waitReply waits for the next response from the client (or timeout/disconnect).
func (w *WebSocketTerminal) waitReply() string {
	select {
	case v := <-w.replyCh:
		return v
	case <-time.After(5 * time.Minute):
		w.mu.Lock()
		w.connected = false
		w.mu.Unlock()
		return ""
	}
}

// --- IOProvider implementation ---

func (w *WebSocketTerminal) Outln(text string, newline bool, color int) {
	msgType := "output"
	if newline {
		text = text + "\n"
	}
	_ = w.send(WSMessage{Type: msgType, Text: text, Color: color})
}

func (w *WebSocketTerminal) ANSICode(code string) {
	// Send raw ANSI codes as a special message; the client can apply them.
	_ = w.send(WSMessage{Type: "ansi", Text: code})
}

func (w *WebSocketTerminal) Cr() {
	_ = w.send(WSMessage{Type: "output", Text: "\n"})
}

func (w *WebSocketTerminal) ClearScreen() {
	_ = w.send(WSMessage{Type: "clear"})
}

func (w *WebSocketTerminal) LettersPrompt(prompt string, allowed string, maxChars int, capitalize bool, showInput bool) string {
	_ = w.send(WSMessage{
		Type:    "prompt_letters",
		Prompt:  prompt,
		Allowed: strings.ToUpper(allowed),
		MaxLen:  maxChars,
	})

	for {
		reply := w.waitReply()
		if !w.IsConnected() {
			return ""
		}
		if capitalize {
			reply = strings.ToUpper(reply)
		}
		if len(reply) == 0 {
			continue
		}
		ch := string(reply[0])
		if capitalize {
			ch = strings.ToUpper(ch)
		}
		if allowed == "" {
			if showInput {
				_ = w.send(WSMessage{Type: "output", Text: ch + "\n"})
			}
			return ch
		}
		if strings.Contains(strings.ToUpper(allowed), strings.ToUpper(ch)) {
			if showInput {
				_ = w.send(WSMessage{Type: "output", Text: ch + "\n"})
			}
			return strings.ToUpper(ch)
		}
		// Invalid choice — ask again.
		_ = w.send(WSMessage{
			Type:    "prompt_letters",
			Prompt:  fmt.Sprintf("Invalid choice. %s", prompt),
			Allowed: strings.ToUpper(allowed),
			MaxLen:  maxChars,
		})
	}
}

func (w *WebSocketTerminal) NumbersPrompt(prompt string, min int, max int) int {
	_ = w.send(WSMessage{
		Type:   "prompt_number",
		Prompt: prompt,
		Min:    min,
		Max:    max,
	})

	for {
		reply := w.waitReply()
		if !w.IsConnected() {
			return min
		}
		n, err := strconv.Atoi(strings.TrimSpace(reply))
		if err != nil || n < min || n > max {
			_ = w.send(WSMessage{
				Type:   "prompt_number",
				Prompt: fmt.Sprintf("Please enter a number between %d and %d. %s", min, max, prompt),
				Min:    min,
				Max:    max,
			})
			continue
		}
		return n
	}
}

func (w *WebSocketTerminal) YesNoQuestion(prompt string) bool {
	_ = w.send(WSMessage{
		Type:    "prompt_yesno",
		Prompt:  prompt,
		Allowed: "YN",
	})

	for {
		reply := w.waitReply()
		if !w.IsConnected() {
			return false
		}
		upper := strings.ToUpper(strings.TrimSpace(reply))
		if upper == "Y" || upper == "YES" {
			return true
		}
		if upper == "N" || upper == "NO" {
			return false
		}
		_ = w.send(WSMessage{
			Type:    "prompt_yesno",
			Prompt:  fmt.Sprintf("Please answer Y or N. %s", prompt),
			Allowed: "YN",
		})
	}
}

func (w *WebSocketTerminal) PausePrompt(prompt string) {
	_ = w.send(WSMessage{
		Type:   "prompt_pause",
		Prompt: prompt,
	})
	// Wait for any reply.
	_ = w.waitReply()
}

func (w *WebSocketTerminal) ReadLine(prompt string, maxLen int) string {
	_ = w.send(WSMessage{
		Type:   "prompt_readline",
		Prompt: prompt,
		MaxLen: maxLen,
	})
	reply := w.waitReply()
	if maxLen > 0 && len(reply) > maxLen {
		reply = reply[:maxLen]
	}
	return reply
}

func (w *WebSocketTerminal) ShowANSIFile(name string) error {
	// Let the client know an ANSI file would be shown here.
	// We send it as plain notification; if the file content were available
	// we'd send it as raw text — skipping for now so as not to send
	// binary ANSI art over the wire.
	_ = w.send(WSMessage{Type: "ansi_file", Text: name})
	return nil
}

func (w *WebSocketTerminal) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}
