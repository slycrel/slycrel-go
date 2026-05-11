import { WhereType } from '../../model/enums.js';
import { TownState } from './town.js';
import { HealerState } from './healer.js';
import { SetupCombatState } from './combat_text.js';

// WildernessState — port of internal/game/states/wilderness.go.
// The combat options (F/M/S) stay as stubs until the combat-text /
// combat-grid states are ported.

export class WildernessState {
  async enter(session) {
    session.character.location = WhereType.TheWildernessMenu;
    session.io.clear();
    await session.io.showAnsiFile('wilderness_menu');
    session.setNext(new WildernessPromptState());
  }
}

export class WildernessPromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println(`  Explorations remaining: ${c.exploration}`, 1);
    io.println("[F]orest, [M]ountain, [S]wamp, [A]gatha's Hut, [Q]uit to Town, [?]Help", 1);
    io.cr();
    const choice = await io.lettersPrompt('Your selection?', 'AFMSQ?');
    io.cr();
    switch (choice) {
      case 'A':
        c.location = WhereType.TheHealersHut;
        session.setNext(new HealerState());
        return;
      case 'F':
      case 'M':
      case 'S':
        c.location = WhereType.TheWildernessCombat;
        session.combatRegion = { F: 'forest', M: 'mountain', S: 'swamp' }[choice];
        session.setNext(new SetupCombatState());
        return;
      case 'Q':
        c.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '?':
        session.setNext(new WildernessState());
        return;
    }
    session.setNext(new WildernessPromptState());
  }
}
