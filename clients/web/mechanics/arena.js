import { randBetween } from './rand.js';

// Convert a character into a Monster record for arena combat.
// Mirrors mechanics/arena.go User2Monster.
export function user2Monster(c) {
  const weapName = c.weapons[0].name || 'fists';
  const weapon2Name = c.weapons[1].name || 'fists';
  const spread = Math.floor(c.speed / 5);
  return {
    name: c.name,
    level: Math.floor((c.thiefLvl + c.mageLvl + c.fighterLvl) / 3),
    monType: 0,
    monsterNum: randBetween(1, 300),
    hitPoints: c.hitPoints,
    offense: Math.floor(c.strength / 2) + (c.weapons[0].strike ?? 0),
    rangeOffense: Math.floor(c.strength / 2) + (c.weapons[1].strike ?? 0),
    range: c.weapons[1].range ?? 0,
    defense: Math.floor((c.dexterity + c.speed) / 3) +
             Math.floor(((c.armor[0]?.defense ?? 0) + (c.armor[1]?.defense ?? 0)) / 2),
    movement: randBetween(c.speed - spread, c.speed + spread),
    disposition: 1,
    experience: Math.floor(c.totalExperience / 10),
    coins: Math.round(c.coinsHand / 1.5),
    attackStr1: weapName,
    attackStr2: weapon2Name,
    defenseStr: weapon2Name,
    defenseActionStr: 'blocks with',
  };
}

export function arenaAppraisal(c) {
  let val = c.totalExperience +
    (c.thiefLvl + c.mageLvl + c.fighterLvl) * 100 +
    c.coinsHand + c.coinsBank;
  val += Math.abs(c.fame) * 100;
  val += Math.abs(c.honor) * 100;
  val += Math.abs(c.faith) * 100;
  val += Math.abs(c.flirt1 + c.flirt2) * 100;
  return val;
}

export function calcOdds(challengerVal, opponentVal) {
  if (challengerVal <= 0 || opponentVal <= 0) return [1, 1];
  const ratio = opponentVal / challengerVal;
  if (ratio < 1 && ratio > 0) {
    return [1, Math.round(challengerVal / opponentVal)];
  }
  return [Math.round(ratio), 1];
}
