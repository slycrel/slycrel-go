package game

import (
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/model"
	"github.com/slycrel/slycrel/internal/store"
)

// Session represents a single player's game session.
// This combines the roles of TSlyPrivates (per-node state) and the
// runtime context that each state handler needs.
type Session struct {
	// Core dependencies
	IO    slyio.IOProvider
	Store store.Store
	SM    *StateMachine

	// Current player
	Character  *model.Character
	Username   string // login username (maps to BBSName)
	TownConfig *model.TownConfig

	// Combat state (mirrors TSlyPrivates fields)
	Monster         *model.Monster
	MonHitPoints    int
	Terrain         *model.TerrainMap
	CellMap         [model.GridRows][model.GridCols]model.CellRec
	UserR, UserC    int // player position on grid
	MonsR, MonsC    int // monster position on grid
	Hit             model.HitType
	UserMove        model.CombatType
	MonMove         model.CombatType
	TextOutln       int          // current line in combat text window (ring index)
	GridStatusMsgs  [5]string    // ring buffer of recent status lines for grid combat
	LastDisposition int

	// Arena state
	CombatName           string
	CombatFileIndex      int64
	CombatForChallenger  bool
	PlayerListCount      int
	CombatBetNum         int

	// Armory state
	ArmoryListCount int
	WeaponToBuy     int
	ArmorToBuy      int

	// Text combat loops
	TextCombatLoop  int
	UserAttackLoop  int
	MonAttackLoop   int

	// Wilderness state
	CombatRegion string // "forest", "mountain", "swamp"

	// Inn state
	DaysRented int
	RentalCost int
}

// NewSession creates a new game session for a player.
func NewSession(io slyio.IOProvider, st store.Store, username string) *Session {
	return &Session{
		IO:       io,
		Store:    st,
		SM:       NewStateMachine(),
		Username: username,
	}
}

// Run is the main game loop. It advances the state machine until
// the player disconnects or quits.
func (s *Session) Run() {
	s.SM.Start("begin")

	for s.IO.IsConnected() && s.SM.Advance() {
		state := s.SM.Current()
		if state == nil {
			break
		}
		state.Enter(s)
	}
}

// SetNextState is a convenience that delegates to the state machine.
func (s *Session) SetNextState(id StateID) {
	s.SM.SetNextState(id)
}

// SetReturnState is a convenience that delegates to the state machine.
func (s *Session) SetReturnState(id StateID) {
	s.SM.SetReturnState(id)
}

// PopReturnState is a convenience that delegates to the state machine.
func (s *Session) PopReturnState() {
	s.SM.PopReturnState()
}

// Quit signals the session to end.
func (s *Session) Quit() {
	s.SM.Quit()
}
