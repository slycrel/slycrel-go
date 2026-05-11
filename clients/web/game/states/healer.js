import { WhereType } from '../../model/enums.js';
import { healCost } from '../../mechanics/leveling.js';
import { saveCharacter } from '../../store/local.js';

// Healer's Hut — port of internal/game/states/healer.go (HealerState and friends).
// Lives in the wilderness, not the town, so the routing back from Q goes
// through the wilderness menu state (TODO once wilderness is ported).

export class HealerState {
  async enter(session) {
    session.character.location = WhereType.TheHealersHut;
    session.io.clear();
    await session.io.showAnsiFile('healer_menu');
    session.setNext(new HealerPromptState());
  }
}

export class HealerPromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println(`You have ${c.hitPoints}/${c.maxHP} Hit Points and ${c.coinsHand} Coins on hand.`, 4);
    io.println('[H]eal All, [N] Heal Amount, [Q]uit, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'HNQ?');
    io.cr();
    switch (choice) {
      case 'H': session.setNext(new AgathaHealsState()); return;
      case 'N': session.setNext(new HealAmountState()); return;
      case 'Q': {
        c.location = WhereType.TheWildernessMenu;
        const { WildernessState } = await import('./wilderness.js');
        session.setNext(new WildernessState());
        return;
      }
      case '?': session.setNext(new HealerState()); return;
    }
    session.setNext(new HealerPromptState());
  }
}

export class AgathaHealsState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const hpNeeded = c.maxHP - c.hitPoints;
    if (hpNeeded <= 0) {
      io.println('You are already at full health!', 3);
      io.cr();
      session.setNext(new HealerPromptState());
      return;
    }
    io.println('Agatha waves her hands strangely above you...', 5);
    io.println('..........', 1);
    io.cr();
    const cost = healCost(c, hpNeeded);
    if (c.coinsHand < cost) {
      io.cr();
      io.println('Her eyes flutter open annoyed at you.', 5);
      io.println("You don't have enough money, THIEF!!!", 6);
      io.cr();
    } else {
      c.coinsHand -= cost;
      c.hitPoints = c.maxHP;
      io.cr();
      io.println(`You have been HEALED! (Cost: ${cost} coins)`, 3);
      io.cr();
    }
    saveCharacter(c);
    session.setNext(new HealerPromptState());
  }
}

export class HealAmountState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const hpNeeded = c.maxHP - c.hitPoints;
    if (hpNeeded <= 0) {
      io.println('You are already at full health!', 3);
      io.cr();
      session.setNext(new HealerPromptState());
      return;
    }
    io.println('How much life would you like healed?', 5);
    const amount = await io.numbersPrompt(':', 0, hpNeeded);
    if (amount <= 0) {
      session.setNext(new HealerPromptState());
      return;
    }
    const cost = healCost(c, amount);
    if (c.coinsHand < cost) {
      io.println('What ARE you trying to do, rip me off???', 6);
      io.cr();
    } else {
      c.coinsHand -= cost;
      c.hitPoints += amount;
      if (c.hitPoints > c.maxHP) c.hitPoints = c.maxHP;
      io.cr();
      io.println(`You have been HEALED! (+${amount} HP, Cost: ${cost} coins)`, 3);
      io.cr();
      saveCharacter(c);
    }
    session.setNext(new HealerPromptState());
  }
}
