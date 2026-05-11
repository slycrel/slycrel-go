import { CombatType } from '../model/enums.js';
import { randBetween } from './rand.js';

// Port of internal/mechanics/combat.go. Player vs monster damage rolls,
// initiative, monster AI, attack/defense modifiers, disposition flavor.

export function rollInitiative(playerMovement, monsterMovement) {
  const diff = playerMovement - monsterMovement;
  return diff + Math.floor(Math.random() * 4) >= 0;
}

export function doDamageUser(c, monster) {
  const avgLevel = Math.floor((c.thiefLvl + c.mageLvl + c.fighterLvl) / 3);
  const weapStrike = c.weapons[0].strike ?? 0;
  let damage;
  if (avgLevel <= 4) {
    const low = Math.round((c.strength * 0.75 + weapStrike) * 1.25);
    const high = Math.round((c.strength / 0.75 + weapStrike) * 1.35);
    damage = randBetween(low, high) - (monster.defense ?? 0);
    if (damage < 2) damage = Math.floor(Math.random() * 3);
  } else {
    const low = Math.round(c.strength / 2 + weapStrike);
    const high = Math.round((c.strength / 2 + weapStrike) * 1.25);
    damage = randBetween(low, high) - (monster.defense ?? 0);
  }
  return damage < 0 ? 0 : damage;
}

export function doDamageMonster(c, monster) {
  const off = monster.offense ?? 0;
  const low = Math.trunc(off * 0.35);
  const high = Math.trunc(off * 1.35);
  const armorDef = c.armor[0]?.defense ?? 0;
  const dmg = randBetween(low, high) - armorDef;
  return dmg < 0 ? 0 : dmg;
}

// Pick a random combat action for the monster. Returns { action, ran }.
export function randMonsterAction(monsterHP, monsterMaxHP) {
  const roll = randBetween(1, 50);
  const wantsToRun = Math.floor(monsterMaxHP / 20) > monsterHP;
  if (roll <= 25) return { action: CombatType.Attack, ran: false };
  if (roll <= 30) return { action: CombatType.Parry, ran: false };
  if (roll <= 37) return { action: CombatType.Block, ran: false };
  if (roll <= 44) return { action: CombatType.Dodge, ran: false };
  return { action: CombatType.Attack, ran: wantsToRun };
}

export function calcAdvanceMoves(attackerLoop, defenderLoop, combatLoop) {
  let moves = 0;
  let loop = combatLoop;
  while (true) {
    loop -= 1;
    if (loop <= 0) break;
    if (attackerLoop >= loop && defenderLoop < loop) {
      moves += 1;
    } else if ((attackerLoop >= loop && defenderLoop >= loop) ||
               (attackerLoop < loop && defenderLoop >= loop)) {
      break;
    }
  }
  return moves;
}

export function applyUserAttackModifier(baseDamage, monMove) {
  if (monMove === CombatType.Block) return Math.floor(baseDamage / 2);
  if (monMove === CombatType.Parry) return Math.round(baseDamage * 0.6);
  return baseDamage;
}

export function applyMonsterAttackModifier(baseDamage, userMove, playerDex, monsterMove) {
  if (userMove === CombatType.Block) return Math.floor(baseDamage / 2);
  if (userMove === CombatType.Parry) return Math.round(baseDamage * 0.6);
  if (userMove === CombatType.Special) return Math.round(baseDamage * 1.75);
  if (userMove === CombatType.Dodge) {
    return randBetween(1, monsterMove) < playerDex ? 0 : baseDamage;
  }
  return baseDamage;
}

// 9-state disposition based on HP%. 1=both good ... 9=both bad. Returns 0
// if either combatant has zero max HP (defensive).
export function getDisposition(userHP, userMax, monHP, monMax) {
  const userPct = userMax > 0 && userHP > 0 ? Math.round(userHP / userMax * 100) : 0;
  const monPct = monMax > 0 && monHP > 0 ? Math.round(monHP / monMax * 100) : 0;
  if (userPct >= 70) {
    if (monPct >= 70) return 1;       // MgoodUgood
    if (monPct >= 35) return 5;       // MokUgood
    return 3;                          // MbadUgood
  } else if (userPct >= 35) {
    if (monPct >= 70) return 7;       // MgoodUok
    if (monPct >= 35) return 4;       // MokUok
    return 8;                          // MbadUok
  } else {
    if (monPct >= 70) return 2;       // MgoodUbad
    if (monPct >= 35) return 6;       // MokUbad
    return 9;                          // MbadUbad
  }
}
