import { WhereType } from '../../model/enums.js';
import { healCost } from '../../mechanics/leveling.js';
import { saveCharacter } from '../../store/remote.js';
import { TownState } from './town.js';

// Herbalist — port of HerbalistState and friends from healer.go.
// Same heal mechanic as the Healer but with slightly different flavor text.

export class HerbalistState {
  async enter(session) {
    session.character.location = WhereType.TheHerbalist;
    session.io.clear();
    await session.io.showAnsiFile('herbalist_menu');
    session.setNext(new HerbalistPromptState());
  }
}

export class HerbalistPromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println(`You have ${c.hitPoints}/${c.maxHP} Hit Points and ${c.coinsHand} Coins on hand.`, 4);
    io.println('[H]eal, [Q]uit, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'HQ?');
    io.cr();
    switch (choice) {
      case 'H': session.setNext(new HerbalistHealState()); return;
      case 'Q':
        c.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '?': session.setNext(new HerbalistState()); return;
    }
    session.setNext(new HerbalistPromptState());
  }
}

export class HerbalistHealState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const hpNeeded = c.maxHP - c.hitPoints;
    if (hpNeeded <= 0) {
      io.println('You are already at full health!', 3);
      io.cr();
      session.setNext(new HerbalistPromptState());
      return;
    }
    io.println('The Herbalist waves her hands strangely above you...', 5);
    io.println('..........', 1);
    io.cr();
    const cost = healCost(c, hpNeeded);
    if (c.coinsHand < cost) {
      io.cr();
      io.println('Her eyes flutter open annoyed at you.', 2);
      io.println("You don't have enough money, THIEF!!!", 6);
      io.cr();
    } else {
      c.coinsHand -= cost;
      c.hitPoints = c.maxHP;
      io.cr();
      io.println(`You have been HEALED!   <da da da da!> (Cost: ${cost} coins)`, 3);
      io.cr();
    }
    await saveCharacter(c);
    session.setNext(new HerbalistPromptState());
  }
}
