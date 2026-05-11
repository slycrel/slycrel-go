import { newDayForUser } from '../../mechanics/resurrection.js';
import { todayDate } from '../../mechanics/rand.js';
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
      session.character.lastOn = todayDate();
      saveCharacter(session.character);
      io.println('A new day dawns... you have been restored.', 3);
      io.cr();
      session.setNext(new BeginState());
    } else {
      // Browser-only: there's no BBS to return to, so quitting the run
      // loop just hangs the page. Drop back to the title instead — the
      // character is still flagged dead, so enter_slycrel will route any
      // re-entry back to this prompt until they pick New Day.
      io.println('Returning to the gateway...', 5);
      io.cr();
      session.setNext(new BeginState());
    }
  }
}
