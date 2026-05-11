import { CharClass } from '../model/enums.js';
import {
  BASE_EXP, STARTING_COINS_HAND, STARTING_COINS_BANK,
  MAX_EXPLORATION, MAX_SPARS,
} from '../model/character.js';
import { randBetween, pick } from './rand.js';

// rollChar mirrors RollChar in internal/game/states/character_create.go.
// Sets level-1 stats and a starting weapon for the chosen class.
export function rollChar(c, charClass) {
  c.charClass = charClass;
  c.fighterLvl = 1;
  c.thiefLvl = 1;
  c.mageLvl = 1;
  c.totalExperience = BASE_EXP;
  c.spendingExperience = BASE_EXP;
  c.coinsHand = STARTING_COINS_HAND;
  c.coinsBank = STARTING_COINS_BANK;
  c.alive = true;
  c.fame = 0; c.honor = 0; c.faith = 0;
  c.flirt1 = 0; c.flirt2 = 0;
  c.totalFights = 0; c.fightsWon = 0;
  c.exploration = MAX_EXPLORATION;
  c.spars = MAX_SPARS;

  switch (charClass) {
    case CharClass.Fighter:
      c.maxHP = randBetween(15, 35);
      c.strength = randBetween(4, 12);
      c.dexterity = randBetween(4, 12);
      c.speed = randBetween(6, 12);
      c.maxPsyche = 1;
      c.movement = randBetween(8, 14);
      c.weapons[0] = { name: 'Short Sword', strike: 3, range: 0, actionStr: 'slash' };
      break;
    case CharClass.Thief:
      c.maxHP = randBetween(13, 26);
      c.strength = randBetween(3, 10);
      c.dexterity = randBetween(8, 20);
      c.speed = randBetween(10, 16);
      c.maxPsyche = 1;
      c.movement = randBetween(10, 18);
      c.weapons[0] = { name: 'Dagger', strike: 2, range: 1, actionStr: 'stab' };
      break;
    case CharClass.Mage:
      c.maxHP = randBetween(4, 17);
      c.strength = randBetween(2, 10);
      c.dexterity = randBetween(4, 14);
      c.speed = randBetween(6, 12);
      c.maxPsyche = randBetween(2, 4);
      c.movement = randBetween(6, 12);
      c.weapons[0] = { name: 'Staff', strike: 1, range: 0, actionStr: 'strike' };
      break;
  }

  c.hitPoints = c.maxHP;
  c.psyche = c.maxPsyche;
}

export function randomClass() {
  return pick([CharClass.Fighter, CharClass.Thief, CharClass.Mage]);
}

// Mirrors NextLevelUp in internal/mechanics/leveling.go.
// f(x) = difficulty * x^3 + f(x-1), base = BASE_EXP at level 1.
export function nextLevelUp(level, difficulty) {
  if (level <= 1) return BASE_EXP;
  let total = BASE_EXP;
  for (let i = 2; i <= level; i++) {
    total += difficulty * i * i * i;
  }
  return total;
}

// Mirrors GiveNewLevel — advances the character one level in their primary
// class and boosts stats.
export function giveNewLevel(c) {
  switch (c.charClass) {
    case CharClass.Fighter:
      c.fighterLvl += 1;
      c.maxHP += randBetween(4, 5 + c.fighterLvl);
      c.hitPoints = c.maxHP;
      c.speed += 1;
      c.strength += randBetween(1, 2);
      c.dexterity += randBetween(0, 1);
      c.fame += 1;
      c.flirt1 += randBetween(0, 2);
      c.flirt2 += randBetween(0, 2);
      break;
    case CharClass.Mage:
      c.mageLvl += 1;
      c.maxHP += randBetween(2, 3 + c.mageLvl);
      c.hitPoints = c.maxHP;
      c.speed += 1;
      c.strength += randBetween(0, 1);
      c.dexterity += randBetween(0, 1);
      c.maxPsyche += randBetween(1, 2);
      c.psyche = c.maxPsyche;
      c.fame += 1;
      c.flirt1 += randBetween(0, 2);
      c.flirt2 += randBetween(0, 2);
      break;
    case CharClass.Thief:
      c.thiefLvl += 1;
      c.maxHP += randBetween(3, 4 + c.thiefLvl);
      c.hitPoints = c.maxHP;
      c.speed += randBetween(1, 2);
      c.strength += randBetween(1, 2);
      c.dexterity += randBetween(1, 2);
      c.fame += 1;
      c.flirt1 += randBetween(0, 2);
      c.flirt2 += randBetween(0, 2);
      break;
  }
  const spread = Math.floor(c.speed / 5);
  c.movement = randBetween(c.speed - spread, c.speed + spread);
}

// Mirrors healCost in internal/game/states/healer.go.
// cost = hpToHeal * floor(avgLevel * 1.25), avgLevel min 1, cost min 1.
export function healCost(c, hpToHeal) {
  let avg = Math.floor((c.thiefLvl + c.mageLvl + c.fighterLvl) / 3);
  if (avg < 2) avg = 1;
  let cost = hpToHeal * Math.trunc(avg * 1.25);
  if (cost < 1 && hpToHeal > 0) cost = 1;
  return cost;
}
