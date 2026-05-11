import { findCharacterByBBSName, getActiveUser, setActiveUser, readNews, clearNews } from '../../store/local.js';
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
