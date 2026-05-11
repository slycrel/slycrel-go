import { newDayForUser } from '../../mechanics/resurrection.js';
import { saveCharacter } from '../../store/local.js';
import { BeginState } from './begin.js';

// DeadState — port of internal/game/states/dead.go.
// Offers a "new day" resurrection (the only way back in a browser-only
// build — we don't model elapsed real-time the way the Go server does).
export class DeadState {
  async enter(session) {
    const { io } = session;
    io.cr();
    const choice = await io.lettersPrompt('[N]ew Day (resurrect & play again) or [Q]uit?', 'NQ');
    io.cr();
    if (choice === 'N') {
      newDayForUser(session.character);
      saveCharacter(session.character);
      io.println('A new day dawns... you have been restored.', 3);
      io.cr();
      session.setNext(new BeginState());
    } else {
      io.println('Now Returning to the BBS', 5);
      io.cr();
      session.quit();
    }
  }
}
