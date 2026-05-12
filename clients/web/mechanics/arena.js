import { randBetween } from './rand.js';
import {
  loadGladiatorFights, loadBets, saveAllGladiatorFights, saveAllBets,
  writeNews,
} from '../store/remote.js';

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

// Resolve any pending gladiator fights this character is involved in
// (as combatant or bettor). Fills out news entries for each resolution
// and mutates the character (rewards/penalties/coins/alive). Called
// during newDay so a fight scheduled yesterday actually pays out.
//
// odds = [challenger, opponent] where the bigger number is the underdog;
// challenger's win probability is odds[1] / (odds[0] + odds[1]).
// Winnings on a successful bet: stake + stake * winnerOdds.
export async function resolveGladiatorFights(c) {
  const fights = await loadGladiatorFights();
  if (!fights.length) return;
  const bets = await loadBets();
  const lcName = (c.name ?? '').toLowerCase();

  // Player's bets keyed by 1-based fight number.
  const myBets = new Map();
  for (const b of bets) {
    if (b.fileName?.toLowerCase() === lcName) myBets.set(b.fightNum, b);
  }

  const remainingFights = [];
  const remainingBets = bets.filter(b => b.fileName?.toLowerCase() !== lcName);

  for (const [idx, fight] of fights.entries()) {
    const fightNum = idx + 1;
    const isChallenger = fight.challenger?.toLowerCase() === lcName;
    const isOpponent = fight.opponent?.toLowerCase() === lcName;
    const involved = isChallenger || isOpponent;
    const myBet = myBets.get(fightNum);

    if (!involved && !myBet) {
      remainingFights.push(fight);
      continue;
    }

    const [oc, oo] = fight.odds;
    const total = oc + oo;
    const roll = randBetween(1, total);
    const challengerWins = roll <= oo;
    const winnerName = challengerWins ? fight.challenger : fight.opponent;
    const loserName = challengerWins ? fight.opponent : fight.challenger;

    if (involved) {
      const playerWon = (isChallenger && challengerWins) || (isOpponent && !challengerWins);
      if (playerWon) {
        const xp = 50, coins = 75;
        c.totalExperience += xp;
        c.spendingExperience += xp;
        c.coinsHand += coins;
        c.fightsWon += 1;
        await writeNews(c.bbsName,
          `Gladiator: You defeated ${loserName} in the arena! +${xp} XP, +${coins} coins.`);
      } else {
        c.totalExperience = Math.max(0, c.totalExperience - 30);
        c.spendingExperience = Math.max(0, c.spendingExperience - 30);
        c.alive = false;
        await writeNews(c.bbsName, `Gladiator: You were slain by ${winnerName} in the arena.`);
      }
      c.totalFights += 1;
    }

    if (myBet) {
      const betWon = (challengerWins && myBet.forChallenger) ||
                     (!challengerWins && !myBet.forChallenger);
      if (betWon) {
        const winnerOdds = challengerWins ? oc : oo;
        const payout = myBet.bet + myBet.bet * winnerOdds;
        c.coinsHand += payout;
        await writeNews(c.bbsName,
          `Bet won: ${fight.challenger} vs ${fight.opponent} — ${winnerName} took it. Payout: ${payout} coins.`);
      } else {
        await writeNews(c.bbsName,
          `Bet lost: ${fight.challenger} vs ${fight.opponent} — ${winnerName} won. You forfeit ${myBet.bet} coins.`);
      }
    }
  }

  await saveAllGladiatorFights(remainingFights);
  await saveAllBets(remainingBets);
}

export function calcOdds(challengerVal, opponentVal) {
  if (challengerVal <= 0 || opponentVal <= 0) return [1, 1];
  const ratio = opponentVal / challengerVal;
  if (ratio < 1 && ratio > 0) {
    return [1, Math.round(challengerVal / opponentVal)];
  }
  return [Math.round(ratio), 1];
}
