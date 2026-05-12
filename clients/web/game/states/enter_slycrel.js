import { findCharacterByBBSName, getActiveUser, login, readNews, saveCharacter } from '../../store/remote.js';
import { newDayForUser } from '../../mechanics/resurrection.js';
import { resolveGladiatorFights } from '../../mechanics/arena.js';
import { runDailyUpkeep } from '../../mechanics/upkeep.js';
import { todayDate } from '../../mechanics/rand.js';
import { CharacterCreateState } from './character_create.js';
import { TownState } from './town.js';
import { BeginState } from './begin.js';

// EnterSlycrelState — port of internal/game/states/begin.go EnterSlycrelState.
// Checks the session cookie via /api/me first. If not logged in, prompts
// for BBS name + password and calls /api/login (which auto-registers new
// BBS names — the supplied password becomes the lock).
export class EnterSlycrelState {
  async enter(session) {
    const { io } = session;

    let user = await getActiveUser();
    if (!user) {
      const name = (await io.textPrompt('Enter your BBS name:', 20)).trim();
      if (!name) {
        session.setNext(new BeginState());
        return;
      }
      const password = await io.passwordPrompt('Password:', 40);
      if (!password) {
        io.cr();
        io.println('Password required.', 6);
        session.setNext(new BeginState());
        return;
      }
      try {
        const result = await login(name, password);
        user = result.bbsName;
        if (result.registered) {
          io.cr();
          io.println('New BBS name claimed.', 3);
        }
      } catch (e) {
        io.cr();
        io.println(`Login failed: ${e.message}`, 6);
        session.setNext(new BeginState());
        return;
      }
    }
    session.username = user;
    io.cr();
    io.println(`-=---  ${user} Entered the realm of Slycrel.`, 1);

    const char = await findCharacterByBBSName(user);
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
      // Resolve any gladiator fights scheduled before today. This may
      // re-kill the character if they lost their match — re-check below.
      await resolveGladiatorFights(char);
      // Shared-world upkeep (inn rent decrement, evictions). Idempotent
      // by inn.lastUpkeep so a second login on the same day is a no-op.
      await runDailyUpkeep(today);
      char.lastOn = today;
      await saveCharacter(char);
      if (!char.alive) {
        io.cr();
        io.println('You died overnight in the gladiator pits...', 6);
        const { DeadState } = await import('./dead.js');
        session.setNext(new DeadState());
        return;
      }
    } else if (!char.alive) {
      io.cr();
      io.println('You are still dead from your last fight.', 6);
      const { DeadState } = await import('./dead.js');
      session.setNext(new DeadState());
      return;
    }

    // readNews atomically fetches + clears the user's news rows.
    const news = await readNews(user);
    if (news) {
      io.cr();
      io.println('=== Daily News ===', 4);
      for (const line of news.split('\n')) io.println(line, 1);
      await io.pausePrompt();
    }

    session.setNext(new TownState());
  }
}
