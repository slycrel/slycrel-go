package game

import "fmt"

// StateMachine manages game flow via registered states and a return stack.
// This faithfully recreates the original's SetNextState/SetReturnState/PopReturnState
// pattern from the Hermes BBS external framework.
type StateMachine struct {
	states      map[StateID]State
	current     StateID
	next        StateID
	returnStack []StateID
	shouldQuit  bool
}

// NewStateMachine creates an empty state machine.
func NewStateMachine() *StateMachine {
	return &StateMachine{
		states: make(map[StateID]State),
	}
}

// Register adds a state to the machine.
func (sm *StateMachine) Register(s State) {
	sm.states[s.ID()] = s
}

// SetNextState sets which state to transition to next.
// Equivalent to the original's CS() / SetNextState().
func (sm *StateMachine) SetNextState(id StateID) {
	sm.next = id
}

// SetReturnState pushes a state onto the return stack.
// When PopReturnState is called, the machine returns to this state.
// Equivalent to the original's RS() / SetReturnState().
func (sm *StateMachine) SetReturnState(id StateID) {
	sm.returnStack = append(sm.returnStack, id)
}

// PopReturnState pops the top of the return stack and sets it as next.
// Equivalent to the original's PopReturnState().
func (sm *StateMachine) PopReturnState() {
	if len(sm.returnStack) == 0 {
		fmt.Println("[WARNING] PopReturnState called with empty stack")
		return
	}
	last := len(sm.returnStack) - 1
	sm.next = sm.returnStack[last]
	sm.returnStack = sm.returnStack[:last]
}

// Current returns the current state, or nil if none.
func (sm *StateMachine) Current() State {
	if s, ok := sm.states[sm.current]; ok {
		return s
	}
	return nil
}

// Advance moves to the next state. Returns false if there is no next state
// or if the machine should quit.
func (sm *StateMachine) Advance() bool {
	if sm.shouldQuit || sm.next == "" {
		return false
	}
	sm.current = sm.next
	sm.next = ""
	return true
}

// Start sets the initial state.
func (sm *StateMachine) Start(id StateID) {
	sm.next = id
}

// Quit signals the state machine to stop.
func (sm *StateMachine) Quit() {
	sm.shouldQuit = true
}

// IsQuitting returns true if the machine has been told to quit.
func (sm *StateMachine) IsQuitting() bool {
	return sm.shouldQuit
}
