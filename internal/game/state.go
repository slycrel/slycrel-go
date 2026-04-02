package game

// StateID identifies a game state. Uses strings for readability
// (the original used integer constants like stTown=200, stArena=600, etc.)
type StateID string

// State represents a single game state - one screen, prompt, or logic step.
// Each ST_* function in the original C++ becomes a State implementation.
type State interface {
	// ID returns this state's unique identifier.
	ID() StateID

	// Enter is called when this state becomes active.
	// The state does its I/O and game logic, then calls session methods
	// (SetNextState, SetReturnState, PopReturnState) to control flow.
	Enter(s *Session)
}
