import { Io } from '../ui/io.js';

// Mirrors internal/game/Session: holds the active character + IO, and drives
// transitions via setNext / setReturn / popReturn (same shape as Go's
// SetNextState / SetReturnState / PopReturnState).
//
// State contract: each state exposes `async enter(session)` and mutates
// session via the transition methods. Returning is for early-exit only —
// the runner reads `_next` to decide what comes next.
export class Session {
  constructor(rootEl) {
    this.io = new Io(rootEl);
    this.character = null;
    this.username = null;
    // Combat scratch — mirrors the Go session fields used by combat_text.
    this.monster = null;
    this.monHitPoints = 0;
    this.userAttackLoop = 0;
    this.monAttackLoop = 0;
    this.textCombatLoop = 0;
    this.hit = 0;
    this.userMove = 0;
    this.monMove = 0;
    this.lastDisposition = 0;
    this.combatRegion = '';
    this._next = null;
    this._returnStack = [];
    this._quit = false;
  }

  setNext(state) { this._next = state; }
  setReturn(state) { this._returnStack.push(state); }
  popReturn() { this._next = this._returnStack.pop() ?? null; }
  quit() { this._quit = true; this._next = null; }

  async run(initialState) {
    this._next = initialState;
    while (this._next && !this._quit) {
      const state = this._next;
      this._next = null;
      await state.enter(this);
    }
  }
}
