package game

import (
	"fmt"
	"sync"

	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/model"
	"github.com/slycrel/slycrel/internal/store"
)

// StructuredIO extends IOProvider with methods to emit structured protocol
// messages at key game state transitions. WebSocketTerminal implements this;
// the plain terminal implementation does not.
type StructuredIO interface {
	slyio.IOProvider
	SetSessionID(id string)
	SendCharacterSnapshot(d slyio.CharacterSnapshotData)
	SendLevelUp(class string, newLevel, maxHP int)
	SendCombatStart(monsterName string, monsterHP, playerHP, playerMaxHP int, mode string)
	SendCombatEnd(won, escaped bool, playerHP int, message string)
}

// GameEngine manages the overall game state and player sessions.
// For the terminal version this runs a single session, but the
// architecture supports multiple concurrent sessions for
// the WebSocket server.
type GameEngine struct {
	Store    store.Store
	DataDir  string
	sessions map[string]*Session
	mu       sync.RWMutex

	// RegisterStates is called to register all game states on a new session's
	// state machine. Set by the caller (main.go) after importing the states package.
	RegisterStates func(sm *StateMachine)
}

// NewGameEngine creates a new game engine.
func NewGameEngine(st store.Store, dataDir string) *GameEngine {
	return &GameEngine{
		Store:    st,
		DataDir:  dataDir,
		sessions: make(map[string]*Session),
	}
}

// GetSession returns the active session for username, or nil if none.
func (e *GameEngine) GetSession(username string) *Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sessions[username]
}

// RunLocalSession creates and runs a single-player session identified by sessionID.
// If io implements StructuredIO, hooks are wired to emit character_snapshot,
// level_up, combat_start, and combat_end protocol messages.
func (e *GameEngine) RunLocalSession(io slyio.IOProvider, username string, sessionID string) error {
	session := NewSession(io, e.Store, username)
	session.SessionID = sessionID

	if e.RegisterStates != nil {
		e.RegisterStates(session.SM)
	}

	cfg, err := e.Store.LoadTownConfig()
	if err != nil {
		return fmt.Errorf("loading town config: %w", err)
	}
	session.TownConfig = cfg

	// Wire structured-protocol hooks when the IO layer supports them.
	if sio, ok := io.(StructuredIO); ok {
		sio.SetSessionID(sessionID)
		session.Hooks = SessionHooks{
			OnCharacterLoaded: func(char *model.Character) {
				sio.SendCharacterSnapshot(charToSnapshotData(char))
			},
			OnLevelUp: func(class string, newLevel, maxHP int) {
				sio.SendLevelUp(class, newLevel, maxHP)
			},
			OnCombatStart: func(monsterName string, monsterHP, playerHP, playerMaxHP int, mode string) {
				sio.SendCombatStart(monsterName, monsterHP, playerHP, playerMaxHP, mode)
			},
			OnCombatEnd: func(won, escaped bool, playerHP int, message string) {
				sio.SendCombatEnd(won, escaped, playerHP, message)
			},
		}
	}

	e.mu.Lock()
	e.sessions[username] = session
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.sessions, username)
		e.mu.Unlock()
	}()

	session.Run()
	return nil
}

// charToSnapshotData converts a model.Character to the transport-layer snapshot type.
func charToSnapshotData(c *model.Character) slyio.CharacterSnapshotData {
	return slyio.CharacterSnapshotData{
		Name:               c.Name,
		Class:              string(c.CharClass),
		Level:              c.Level(),
		FighterLvl:         c.FighterLvl,
		ThiefLvl:           c.ThiefLvl,
		MageLvl:            c.MageLvl,
		HitPoints:          c.HitPoints,
		MaxHP:              c.MaxHP,
		Strength:           c.Strength,
		Dexterity:          c.Dexterity,
		Speed:              c.Speed,
		Psyche:             c.Psyche,
		MaxPsyche:          c.MaxPsyche,
		CoinsHand:          c.CoinsHand,
		CoinsBank:          c.CoinsBank,
		TotalExperience:    c.TotalExperience,
		SpendingExperience: c.SpendingExperience,
		Location:           c.Location.String(),
		Alive:              c.Alive,
	}
}
