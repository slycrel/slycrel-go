import { WhereType } from '../../model/enums.js';
import { saveCharacter } from '../../store/local.js';
import { BeginState } from './begin.js';
import { BankState } from './bank.js';

// Mapping of the town menu's letter keys to human-readable location names,
// used by the stub message until each location is ported.
const LOCATION_BY_KEY = {
  A: { name: 'Arena',          where: WhereType.TheArenaMenu },
  B: { name: 'Bank',           where: WhereType.TheBank },
  C: { name: 'Common Guild',   where: WhereType.TheCommonGuild },
  H: { name: 'Herbalist',      where: WhereType.TheHerbalist },
  I: { name: 'Inn',            where: WhereType.TheInn },
  J: { name: 'Jail',           where: WhereType.TheJail },
  L: { name: 'Armory',         where: WhereType.TheArmory },
  S: { name: 'Blacksmith',     where: WhereType.TheBlacksmiths },
  T: { name: 'Tavern',         where: WhereType.TheTavern },
  W: { name: 'Wilderness',     where: WhereType.TheWildernessMenu },
  '@': { name: 'Tower',        where: WhereType.TheTower },
};

// TownState — renders main_menu.ans (the town map ANSI screen).
// Mirrors internal/game/states/town.go TownState.
export class TownState {
  async enter(session) {
    if (session.character) session.character.location = WhereType.TheTown;
    session.io.clear();
    await session.io.showAnsiFile('main_menu');
    session.setNext(new TownPromptState());
  }
}

// TownPromptState — reads the menu choice and routes (or stubs).
// Mirrors internal/game/states/town.go TownPromptState.
export class TownPromptState {
  async enter(session) {
    const { io } = session;
    io.println('[A,B,C,H,I,J,L,S,T,V,W,X,@,?]', 1);
    io.cr();

    const choice = await io.lettersPrompt('Your Choice?', 'ABCHIJLSTVWX@?');
    io.cr();

    if (choice === '?') {
      session.setNext(new TownState());
      return;
    }
    if (choice === 'V') {
      // View character — defer to a real port once view_character is callable
      // from the town menu. For now, stub like the rest of the locations.
      io.println('(stub) character sheet from town menu not yet implemented.', 6);
      await io.pausePrompt();
      session.setNext(new TownPromptState());
      return;
    }
    if (choice === 'X') {
      io.println('Saving and returning to the gateway...', 5);
      saveCharacter(session.character);
      await io.pausePrompt();
      session.setNext(new BeginState());
      return;
    }

    if (choice === 'B') {
      session.setNext(new BankState());
      return;
    }

    const loc = LOCATION_BY_KEY[choice];
    if (loc) {
      session.character.location = loc.where;
      io.println(`(stub) ${loc.name} not yet implemented.`, 6);
      await io.pausePrompt();
      session.setNext(new TownPromptState());
      return;
    }

    session.setNext(new TownPromptState());
  }
}
