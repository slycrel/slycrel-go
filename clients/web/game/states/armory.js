import { WhereType } from '../../model/enums.js';
import { loadWeapons, loadArmor, saveCharacter } from '../../store/local.js';
import { TownState } from './town.js';

const ARMOR_CLASS_NAMES = ['Body', 'Shield', 'Full Body'];

// Armory — port of internal/game/states/armory.go.
// Loads weapons.json and armor.json on demand and lets the player buy,
// sell, or swap their primary/off-hand weapon slot.

export class ArmoryState {
  async enter(session) {
    session.character.location = WhereType.TheArmory;
    session.io.clear();
    await session.io.showAnsiFile('armory_menu');
    session.setNext(new ArmoryPromptState());
  }
}

export class ArmoryPromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println(`  Coins on hand: ${c.coinsHand}`, 4);
    io.println('[W]eapons, [A]rmor, [S]ell Weapon, [G] Sell Armor, [M]ain/Off Switch, [Q]uit, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'WASGMQ?');
    io.cr();
    switch (choice) {
      case 'W': session.setNext(new BuyWeaponState()); return;
      case 'A': session.setNext(new BuyArmorState()); return;
      case 'S': session.setNext(new SellWeaponState()); return;
      case 'G': session.setNext(new SellArmorState()); return;
      case 'M':
        io.println('You switch your weapons.', 1);
        io.cr();
        [c.weapons[0], c.weapons[1]] = [c.weapons[1], c.weapons[0]];
        saveCharacter(c);
        session.setNext(new ArmoryPromptState());
        return;
      case 'Q':
        c.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '?': session.setNext(new ArmoryState()); return;
    }
    session.setNext(new ArmoryPromptState());
  }
}

export class BuyWeaponState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const weapons = await loadWeapons();
    if (!weapons.length) {
      io.println('No weapons available!', 6);
      await io.pausePrompt();
      session.setNext(new ArmoryPromptState());
      return;
    }
    io.cr();
    io.println('=--=-- Available Weapons ---=--=', 3);
    io.cr();
    weapons.forEach((w, i) => {
      const affordable = w.cost <= c.coinsHand ? '*' : ' ';
      const rangeStr = w.range > 0 ? ` Range:${w.range}` : '';
      const idx = String(i + 1).padStart(2);
      const name = (w.name || '').padEnd(18);
      const strike = String(w.strike).padEnd(3);
      const cost = String(w.cost).padEnd(4);
      io.println(` ${affordable}${idx}) ${name} Strike:${strike}${rangeStr}  Cost:${cost}`, 1);
    });
    io.cr();
    io.println('  (* = you can afford)', 4);
    io.cr();
    const num = await io.numbersPrompt('Choose Your Weapon (0=Quit):', 0, weapons.length);
    if (num === 0) {
      session.setNext(new ArmoryPromptState());
      return;
    }
    const weapon = weapons[num - 1];
    if (!await io.yesNoQuestion(`Buy ${weapon.name} for ${weapon.cost} coins?`)) {
      io.println("FINE, I didn't want your stupid weapon anyway...", 1);
      io.cr();
      session.setNext(new ArmoryPromptState());
      return;
    }
    if (c.coinsHand < weapon.cost) {
      io.cr();
      io.println("Sorry, I'm not giving these away. Come back when you get more money.", 6);
      io.cr();
      await io.pausePrompt('--More--');
      session.setNext(new ArmoryPromptState());
      return;
    }
    io.cr();
    io.println(`  Weapon 1: ${c.weapons[0].name || 'Empty'}`, 1);
    io.println(`  Weapon 2: ${c.weapons[1].name || 'Empty'}`, 1);
    io.cr();
    const slot = await io.numbersPrompt('Which weapon slot to replace? (1 or 2):', 1, 2);
    c.weapons[slot - 1] = { ...weapon };
    c.coinsHand -= weapon.cost;
    saveCharacter(c);
    io.cr();
    io.println(`Okay, you buy the spiffy ${weapon.name}.`, 3);
    await io.pausePrompt('--More--');
    session.setNext(new ArmoryPromptState());
  }
}

export class BuyArmorState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const armors = await loadArmor();
    if (!armors.length) {
      io.println('No armor available!', 6);
      await io.pausePrompt();
      session.setNext(new ArmoryPromptState());
      return;
    }
    io.cr();
    io.println('=--=-- Available Armor ---=--=', 3);
    io.cr();
    armors.forEach((a, i) => {
      const affordable = a.cost <= c.coinsHand ? '*' : ' ';
      const cn = ARMOR_CLASS_NAMES[a.class] ?? 'Body';
      const idx = String(i + 1).padStart(2);
      const name = (a.name || '').padEnd(18);
      const def = String(a.defense).padEnd(3);
      const type = cn.padEnd(9);
      const cost = String(a.cost).padEnd(4);
      io.println(` ${affordable}${idx}) ${name} Def:${def} Type:${type} Cost:${cost}`, 1);
    });
    io.cr();
    io.println('  (* = you can afford)', 4);
    io.cr();
    const num = await io.numbersPrompt('Choose Your Armor (0=Quit):', 0, armors.length);
    if (num === 0) {
      session.setNext(new ArmoryPromptState());
      return;
    }
    const armor = armors[num - 1];
    if (!await io.yesNoQuestion(`Buy ${armor.name} for ${armor.cost} coins?`)) {
      io.println('Okay, your loss. Have fun dying.', 1);
      io.cr();
      session.setNext(new ArmoryPromptState());
      return;
    }
    if (c.coinsHand < armor.cost) {
      io.cr();
      io.println("Sorry, I'm not giving these away. Come back when you get more money.", 6);
      io.cr();
      session.setNext(new ArmoryPromptState());
      return;
    }
    c.armor[0] = { ...armor };
    c.coinsHand -= armor.cost;
    saveCharacter(c);
    io.cr();
    io.println(`Okay, you buy the spiffy ${armor.name}.`, 3);
    await io.pausePrompt('--More--');
    session.setNext(new ArmoryPromptState());
  }
}

export class SellWeaponState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const w = c.weapons[0];
    const sellPrice = Math.floor((w.cost ?? 0) / 2);
    if (!w.name || w.name === 'Hands') {
      io.cr();
      io.println("You can't sell your hands. Idiot...", 6);
      io.cr();
      await io.pausePrompt('-More-');
      session.setNext(new ArmoryPromptState());
      return;
    }
    if ((w.cost ?? 0) <= 0) {
      io.cr();
      io.println("Sorry, I don't buy junk.", 6);
      io.cr();
      await io.pausePrompt('-More-');
      session.setNext(new ArmoryPromptState());
      return;
    }
    if (!await io.yesNoQuestion(`Sell your ${w.name} for ${sellPrice} coins?`)) {
      io.cr();
      io.println("Okay, you jerk, I didn't want it anyway.", 3);
      io.cr();
      await io.pausePrompt('-More-');
      session.setNext(new ArmoryPromptState());
      return;
    }
    c.coinsHand += sellPrice;
    c.weapons[0] = { name: 'Hands', strike: 1, range: 0, actionStr: 'punch', cost: 0 };
    saveCharacter(c);
    io.cr();
    io.println('Okay, thanks. Come back anytime.', 1);
    io.cr();
    await io.pausePrompt('-More-');
    session.setNext(new ArmoryPromptState());
  }
}

export class SellArmorState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const a = c.armor[0];
    const sellPrice = Math.floor((a.cost ?? 0) / 2);
    if (!a.name || (a.cost ?? 0) <= 0) {
      io.cr();
      io.println("Sorry, I don't buy junk.", 6);
      io.cr();
      await io.pausePrompt('-More-');
      session.setNext(new ArmoryPromptState());
      return;
    }
    if (!await io.yesNoQuestion(`Sell your ${a.name} for ${sellPrice} coins?`)) {
      io.cr();
      io.println("Okay, you jerk, I didn't want it anyway.", 3);
      io.cr();
      await io.pausePrompt('-More-');
      session.setNext(new ArmoryPromptState());
      return;
    }
    c.coinsHand += sellPrice;
    c.armor[0] = { name: '', defense: 0, cost: 0 };
    saveCharacter(c);
    io.cr();
    io.println('Okay, thanks. Come back anytime.', 1);
    io.cr();
    await io.pausePrompt('-More-');
    session.setNext(new ArmoryPromptState());
  }
}
