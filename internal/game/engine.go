package game

import (
	"fmt"
	"sync"

	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/store"
)

// GameEngine manages the overall game state and player sessions.
// For the terminal version this runs a single session, but the
// architecture supports multiple concurrent sessions for future
// telnet/WebSocket server modes.
type GameEngine struct {
	Store    store.Store
	DataDir  string
	sessions map[string]*Session
	mu       sync.RWMutex

	// registerStates is called to register all game states on a new session's
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

// RunLocalSession creates and runs a single-player terminal session.
// This is the entry point for the terminal app.
func (e *GameEngine) RunLocalSession(io slyio.IOProvider, username string) error {
	session := NewSession(io, e.Store, username)

	if e.RegisterStates != nil {
		e.RegisterStates(session.SM)
	}

	// Load town config
	cfg, err := e.Store.LoadTownConfig()
	if err != nil {
		return fmt.Errorf("loading town config: %w", err)
	}
	session.TownConfig = cfg

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
