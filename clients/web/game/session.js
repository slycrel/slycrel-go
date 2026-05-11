import { Io } from '../ui/io.js';

// Session mirrors internal/game/Session: holds the current character, the IO
// abstraction, and drives the state machine. States are objects with an
// async `enter(session)` method; transitions are returned (or null to stop).
export class Session {
  constructor(rootEl) {
    this.io = new Io(rootEl);
    this.character = null;
    this.username = null;
  }

  async run(initialState) {
    let state = initialState;
    while (state) {
      state = (await state.enter(this)) ?? null;
    }
  }
}
