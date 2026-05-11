import { findCharacterByBBSName, getActiveUser, setActiveUser, readNews, clearNews, saveCharacter } from '../../store/local.js';
import { newDayForUser } from '../../mechanics/resurrection.js';
import { todayDate } from '../../mechanics/rand.js';
import { CharacterCreateState } from './character_create.js';
import { TownState } from './town.js';
import { BeginState } from './begin.js';

// EnterSlycrelState — port of internal/game/states/begin.go EnterSlycrelState.
// Loads the active user's character or routes to creation.
export class EnterSlycrelState {
  async enter(session) {
    const { io } = session;

    let user = getActiveUser();
    if (!user) {
      user = (await io.textPrompt('Enter your BBS name:', 20)).trim();
      if (!user) {
        session.setNext(new BeginState());
        return;
      }
      setActiveUser(user);
    }
    session.username = user;
    io.cr();
    io.println(`-=---  ${user} Entered the realm of Slycrel.`, 1);

    const char = findCharacterByBBSName(user);
    const needsCreation = !char || !char.name || !char.maxHP;
    if (needsCreation) {
      io.cr();
      io.println('No character found. Creating a new one...', 3);
      session.character = char ?? null;
      session.setNext(new CharacterCreateState());
      return;
    }

    session.character = char;

    // New-day check mirrors enter_slycrel.go: if lastOn is older than
    // today, refresh daily limits (and resurrect a dead character).
    // If still dead within the same day, the player has to wait until
    // tomorrow or use the resurrect prompt manually.
    const today = todayDate();
    if (char.lastOn < today) {
      if (!char.alive) {
        io.cr();
        io.println('A new day dawns... you have been restored.', 3);
      }
      newDayForUser(char);
      char.lastOn = today;
      saveCharacter(char);
    } else if (!char.alive) {
      io.cr();
      io.println('You are still dead from your last fight.', 6);
      const { DeadState } = await import('./dead.js');
      session.setNext(new DeadState());
      return;
    }

    // Show pending arena mail (e.g. "you've been slaughtered by X").
    const news = readNews(user);
    if (news) {
      io.cr();
      io.println('=== Daily News ===', 4);
      for (const line of news.split('\n')) io.println(line, 1);
      clearNews(user);
      await io.pausePrompt();
    }

    session.setNext(new TownState());
  }
}
