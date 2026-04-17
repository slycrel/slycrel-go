package io

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage is the JSON envelope for all WebSocket messages.
type WSMessage struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Prompt    string          `json:"prompt,omitempty"`
	Allowed   string          `json:"allowed,omitempty"`
	Min       int             `json:"min,omitempty"`
	Max       int             `json:"max,omitempty"`
	MaxLen    int             `json:"maxLen,omitempty"`
	Color     int             `json:"color,omitempty"`
	SessionID string          `json:"sessionId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// WSClientMessage is a message from the browser back to the server.
type WSClientMessage struct {
	Type   string `json:"type"`
	Value  string `json:"value"`  // for input and most typed commands
	Action string `json:"action"` // alias for Value used by some typed command messages
}

// Structured payload types for server→client messages.

type SessionInitPayload struct {
	SessionID string `json:"sessionId"`
	Username  string `json:"username"`
}

type SessionRestoredPayload struct {
	SessionID string `json:"sessionId"`
	Username  string `json:"username"`
}

type CharacterSnapshotPayload struct {
	Name               string `json:"name"`
	Class              string `json:"class"`
	Level              int    `json:"level"`
	FighterLvl         int    `json:"fighterLvl"`
	ThiefLvl           int    `json:"thiefLvl"`
	MageLvl            int    `json:"mageLvl"`
	HitPoints          int    `json:"hitPoints"`
	MaxHP              int    `json:"maxHp"`
	Strength           int    `json:"strength"`
	Dexterity          int    `json:"dexterity"`
	Speed              int    `json:"speed"`
	Psyche             int    `json:"psyche"`
	MaxPsyche          int    `json:"maxPsyche"`
	CoinsHand          int64  `json:"coinsHand"`
	CoinsBank          int64  `json:"coinsBank"`
	TotalExperience    int64  `json:"totalExperience"`
	SpendingExperience int64  `json:"spendingExperience"`
	Location           string `json:"location"`
	Alive              bool   `json:"alive"`
}

type LevelUpPayload struct {
	Class    string `json:"class"`
	NewLevel int    `json:"newLevel"`
	MaxHP    int    `json:"maxHp"`
}

type CombatStartPayload struct {
	MonsterName string `json:"monsterName"`
	MonsterHP   int    `json:"monsterHp"`
	PlayerHP    int    `json:"playerHp"`
	PlayerMaxHP int    `json:"playerMaxHp"`
	Mode        string `json:"mode"` // "text" or "grid"
}

type CombatEndPayload struct {
	Won      bool   `json:"won"`
	Escaped  bool   `json:"escaped"`
	PlayerHP int    `json:"playerHp"`
	Message  string `json:"message"`
}

// CharacterSnapshotData carries character stats for SendCharacterSnapshot.
// Defined here (not in model) to keep the IO package self-contained.
type CharacterSnapshotData struct {
	Name               string
	Class              string
	Level              int
	FighterLvl         int
	ThiefLvl           int
	MageLvl            int
	HitPoints          int
	MaxHP              int
	Strength           int
	Dexterity          int
	Speed              int
	Psyche             int
	MaxPsyche          int
	CoinsHand          int64
	CoinsBank          int64
	TotalExperience    int64
	SpendingExperience int64
	Location           string
	Alive              bool
}

// WebSocketTerminal implements IOProvider over a WebSocket connection.
// The game engine calls the same IOProvider methods as the terminal version;
// this implementation serialises them to JSON and exchanges them with the
// browser client over the WebSocket.
type WebSocketTerminal struct {
	conn      *websocket.Conn
	mu        sync.Mutex // protects conn writes and connected flag
	replyCh   chan string // input responses normalised from all client command types
	connected bool
	dataDir   string
	sessionID string
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

// SetSessionID stores the session ID included in outgoing messages.
func (w *WebSocketTerminal) SetSessionID(id string) {
	w.mu.Lock()
	w.sessionID = id
	w.mu.Unlock()
}

// SwapConn atomically replaces the underlying WebSocket connection (reconnect path).
// The old connection is closed; a new read loop is started on the replacement conn.
func (w *WebSocketTerminal) SwapConn(conn *websocket.Conn) {
	w.mu.Lock()
	old := w.conn
	w.conn = conn
	w.connected = true
	w.mu.Unlock()

	if old != nil {
		_ = old.Close()
	}
	go w.readLoop()
}

// readLoop pumps messages from the browser into replyCh, normalising all typed
// command messages (combat_action, grid_move, shop_buy, arena_action) to their
// string value so existing LettersPrompt/NumbersPrompt handlers work unchanged.
func (w *WebSocketTerminal) readLoop() {
	defer func() {
		w.mu.Lock()
		w.connected = false
		w.mu.Unlock()
		// Unblock any waiting prompt.
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

		// Normalise typed command messages to a string pushed to replyCh.
		// The state machine's LettersPrompt/NumbersPrompt handlers are unaware of
		// which structured type the client used — they just read from replyCh.
		var reply string
		switch msg.Type {
		case "input":
			reply = msg.Value
		case "combat_action", "grid_move", "shop_buy", "arena_action", "arena_challenge_request":
			reply = msg.Value
			if reply == "" {
				reply = msg.Action
			}
		default:
			continue
		}

		select {
		case w.replyCh <- reply:
		default:
			// Drop if nobody is waiting.
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

// sendWithData sends a message whose payload is a marshalled struct.
func (w *WebSocketTerminal) sendWithData(msgType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	return w.send(WSMessage{Type: msgType, Data: json.RawMessage(raw)})
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

// --- Structured protocol helpers (called by server wiring, not game states) ---

// SendSessionInit sends the session_init message when a new session starts.
func (w *WebSocketTerminal) SendSessionInit(username, sessionID string) {
	_ = w.sendWithData("session_init", SessionInitPayload{
		SessionID: sessionID,
		Username:  username,
	})
}

// SendSessionRestored sends the session_restored message on reconnect.
func (w *WebSocketTerminal) SendSessionRestored(username, sessionID string) {
	_ = w.sendWithData("session_restored", SessionRestoredPayload{
		SessionID: sessionID,
		Username:  username,
	})
}

// SendSessionEnd signals the client that the session has ended cleanly.
func (w *WebSocketTerminal) SendSessionEnd() {
	_ = w.send(WSMessage{Type: "session_end"})
}

// SendCharacterSnapshot sends a full character state snapshot to the client.
func (w *WebSocketTerminal) SendCharacterSnapshot(d CharacterSnapshotData) {
	_ = w.sendWithData("character_snapshot", CharacterSnapshotPayload{
		Name:               d.Name,
		Class:              d.Class,
		Level:              d.Level,
		FighterLvl:         d.FighterLvl,
		ThiefLvl:           d.ThiefLvl,
		MageLvl:            d.MageLvl,
		HitPoints:          d.HitPoints,
		MaxHP:              d.MaxHP,
		Strength:           d.Strength,
		Dexterity:          d.Dexterity,
		Speed:              d.Speed,
		Psyche:             d.Psyche,
		MaxPsyche:          d.MaxPsyche,
		CoinsHand:          d.CoinsHand,
		CoinsBank:          d.CoinsBank,
		TotalExperience:    d.TotalExperience,
		SpendingExperience: d.SpendingExperience,
		Location:           d.Location,
		Alive:              d.Alive,
	})
}

// SendLevelUp notifies the client of a level-up event.
func (w *WebSocketTerminal) SendLevelUp(class string, newLevel, maxHP int) {
	_ = w.sendWithData("level_up", LevelUpPayload{
		Class:    class,
		NewLevel: newLevel,
		MaxHP:    maxHP,
	})
}

// SendCombatStart notifies the client that combat has begun.
func (w *WebSocketTerminal) SendCombatStart(monsterName string, monsterHP, playerHP, playerMaxHP int, mode string) {
	_ = w.sendWithData("combat_start", CombatStartPayload{
		MonsterName: monsterName,
		MonsterHP:   monsterHP,
		PlayerHP:    playerHP,
		PlayerMaxHP: playerMaxHP,
		Mode:        mode,
	})
}

// SendCombatEnd notifies the client that combat has ended.
func (w *WebSocketTerminal) SendCombatEnd(won, escaped bool, playerHP int, message string) {
	_ = w.sendWithData("combat_end", CombatEndPayload{
		Won:      won,
		Escaped:  escaped,
		PlayerHP: playerHP,
		Message:  message,
	})
}

// --- IOProvider implementation ---

func (w *WebSocketTerminal) Outln(text string, newline bool, color int) {
	if newline {
		text = text + "\n"
	}
	_ = w.send(WSMessage{Type: "output", Text: text, Color: color})
}

func (w *WebSocketTerminal) ANSICode(code string) {
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
	path := filepath.Join(w.dataDir, "ansi", name+".ans")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ws: ShowANSIFile %q: %v", name, err)
		_ = w.send(WSMessage{Type: "ansi_file", Text: name})
		return err
	}
	_ = w.send(WSMessage{Type: "ansi_file", Text: string(data)})
	return nil
}

func (w *WebSocketTerminal) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}
