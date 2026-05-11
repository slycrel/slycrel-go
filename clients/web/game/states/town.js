import { WhereType } from '../../model/enums.js';
import { saveCharacter } from '../../store/local.js';
import { BeginState } from './begin.js';
import { BankState } from './bank.js';
import { JailState } from './jail.js';
import { TowerState } from './tower.js';
import { BlacksmithState } from './blacksmith.js';
import { ViewCharacterState } from './view_character.js';
import { HerbalistState } from './herbalist.js';
import { CommonGuildState } from './common_guild.js';
import { TavernState } from './tavern.js';
import { ArmoryState } from './armory.js';
import { InnState } from './inn.js';
import { WildernessState } from './wilderness.js';

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
      session.setReturn(new TownPromptState());
      session.setNext(new ViewCharacterState());
      return;
    }
    if (choice === 'X') {
      io.println('Saving and returning to the gateway...', 5);
      saveCharacter(session.character);
      await io.pausePrompt();
      session.setNext(new BeginState());
      return;
    }

    const SUB_STATES = {
      B: () => new BankState(),
      J: () => new JailState(),
      S: () => new BlacksmithState(),
      '@': () => new TowerState(),
      H: () => new HerbalistState(),
      C: () => new CommonGuildState(),
      T: () => new TavernState(),
      L: () => new ArmoryState(),
      I: () => new InnState(),
      W: () => new WildernessState(),
    };
    if (SUB_STATES[choice]) {
      session.setNext(SUB_STATES[choice]());
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
