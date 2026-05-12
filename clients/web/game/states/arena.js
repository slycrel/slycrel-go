import { WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import { user2Monster, arenaAppraisal, calcOdds } from '../../mechanics/arena.js';
import {
  saveCharacter, findCharacterByName, listCharacters,
  loadGladiatorFights, saveGladiatorFight, saveBet, writeNews,
} from '../../store/remote.js';
import { TownState } from './town.js';
import { ViewCharacterState } from './view_character.js';
import { GridCombatSetupState } from './combat_grid.js';

export class ArenaState {
  async enter(session) {
    session.character.location = WhereType.TheArenaMenu;
    session.io.clear();
    await session.io.showAnsiFile('arena_menu');
    session.setNext(new ArenaPromptState());
  }
}

export class ArenaPromptState {
  async enter(session) {
    const { io } = session;
    io.println('[B]et, [C]hallenge, [G]ladiator Fight, [L]ist Players, [Q]uit, [V]iew Stats, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'BCGLQV?');
    io.cr();
    switch (choice) {
      case 'B': session.setNext(new ArenaBetState()); return;
      case 'C': session.setNext(new ArenaChallengeState()); return;
      case 'G': session.setNext(new ArenaGladiatorState()); return;
      case 'L': session.setNext(new ArenaListPlayersState()); return;
      case 'Q':
        session.character.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case 'V':
        session.setReturn(new ArenaPromptState());
        session.setNext(new ViewCharacterState());
        return;
      case '?': session.setNext(new ArenaState()); return;
    }
    session.setNext(new ArenaPromptState());
  }
}

export class ArenaChallengeState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    if (c.spars < 1) {
      io.println("Sorry Buddy, but you've fought All that you can Today.", 2);
      session.setNext(new ArenaPromptState());
      return;
    }
    const name = (await io.textPrompt('Player to Challenge:', 27)).trim();
    if (!name) {
      session.setNext(new ArenaPromptState());
      return;
    }
    const opponent = await findCharacterByName(name);
    if (!opponent) {
      io.println('Player Not Found', 6);
      session.setNext(new ArenaPromptState());
      return;
    }
    if (!opponent.alive) {
      io.println("That Player is Dead, you can't fight him!", 1);
      session.setNext(new ArenaPromptState());
      return;
    }
    if (opponent.name.toLowerCase() === c.name.toLowerCase()) {
      io.println('You can not challenge yourself. Fool!', 3);
      session.setNext(new ArenaPromptState());
      return;
    }

    const monster = user2Monster(opponent);
    session.monster = monster;
    session.monHitPoints = monster.hitPoints;
    session.combatName = opponent.name;
    c.location = WhereType.TheArenaCombat;
    c.spars -= 1;
    c.totalFights += 1;
    const spd = Math.max(c.speed, 5);
    const spread = Math.floor(spd / 5);
    c.movement = randBetween(spd - spread, spd + spread);

    session.combatRegion = 'arena';
    io.clear();
    io.println(`You face ${opponent.name} in the Arena!`, 6);
    io.cr();
    await saveCharacter(c);
    session.setNext(new GridCombatSetupState());
  }
}

export class ArenaGladiatorState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    if (c.spars < 1) {
      io.println("Sorry Buddy, but you've fought All that you can Today.", 2);
      session.setNext(new ArenaPromptState());
      return;
    }
    const name = (await io.textPrompt('Player to Challenge:', 27)).trim();
    if (!name) {
      session.setNext(new ArenaPromptState());
      return;
    }
    const opponent = await findCharacterByName(name);
    if (!opponent) {
      io.println('Player Not Found', 6);
      session.setNext(new ArenaPromptState());
      return;
    }
    if (!opponent.alive) {
      io.println("That Player is Dead, you can't fight him!", 1);
      session.setNext(new ArenaPromptState());
      return;
    }
    if (opponent.name.toLowerCase() === c.name.toLowerCase()) {
      io.println('You can not challenge yourself. Fool!', 3);
      session.setNext(new ArenaPromptState());
      return;
    }
    if (!await io.yesNoQuestion('Are you sure you want to go through with this?')) {
      io.cr();
      io.println('You regretfully decide not to go through with this.', 1);
      session.setNext(new ArenaPromptState());
      return;
    }

    const challengerVal = arenaAppraisal(c);
    const opponentVal = arenaAppraisal(opponent);
    const odds = calcOdds(challengerVal, opponentVal);
    await saveGladiatorFight({ challenger: c.name, opponent: opponent.name, odds });
    c.spars -= 1;
    await saveCharacter(c);
    io.println(`You have entered into the Gladiator competition against ${opponent.name} (${opponent.bbsName})!`, 1);
    io.println(`Odds (Challenger:Opponent) = ${odds[0]}:${odds[1]}`, 4);
    io.cr();
    session.setNext(new ArenaPromptState());
  }
}

export class ArenaBetState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const fights = await loadGladiatorFights();
    if (!fights.length) {
      io.println('Sorry, there are no matches to bet on at this time.', 3);
      io.cr();
      session.setNext(new ArenaPromptState());
      return;
    }
    io.cr();
    fights.forEach((f, i) => {
      io.println(`${i + 1}) ${f.challenger} vs ${f.opponent}  (Odds ${f.odds[0]}:${f.odds[1]})`, 1);
    });
    io.cr();
    const fightNum = await io.numbersPrompt('Which fight? (0=Quit):', 0, fights.length);
    if (fightNum === 0) {
      session.setNext(new ArenaPromptState());
      return;
    }
    const fight = fights[fightNum - 1];
    if (fight.challenger.toLowerCase() === c.name.toLowerCase() ||
        fight.opponent.toLowerCase() === c.name.toLowerCase()) {
      io.println("Sorry, you can't bet when you're already busy fighting.", 3);
      session.setNext(new ArenaPromptState());
      return;
    }
    const choice = await io.lettersPrompt('Bet [F]or or [A]gainst the challenger? (Q=Quit)', 'FAQ');
    if (choice === 'Q') {
      session.setNext(new ArenaPromptState());
      return;
    }
    const forChallenger = choice === 'F';
    const maxBet = Math.min(1000, c.coinsHand);
    if (maxBet <= 0) {
      io.println("You don't have any coins to bet!", 6);
      session.setNext(new ArenaPromptState());
      return;
    }
    const amount = await io.numbersPrompt(`How much do you wish to bet? [1..${maxBet}]:`, 1, maxBet);
    if (amount > c.coinsHand) {
      io.println("Sorry, you don't have enough cash.", 1);
      io.cr();
      session.setNext(new ArenaPromptState());
      return;
    }
    c.coinsHand -= amount;
    await saveBet({
      fightNum, fileName: c.name, bet: amount,
      challenger: fight.challenger, opponent: fight.opponent, forChallenger,
    });
    await saveCharacter(c);
    io.println('Okay, you now are in the pool!', 2);
    io.cr();
    await io.pausePrompt('-Any Key-');
    session.setNext(new ArenaPromptState());
  }
}

export class ArenaListPlayersState {
  async enter(session) {
    const { io } = session;
    const chars = await listCharacters();
    if (!chars.length) {
      io.println('No players found.', 1);
      session.setNext(new ArenaPromptState());
      return;
    }
    io.cr();
    io.println(' Name                      Score    Alive  Class', 4);
    io.println(' ------------------------------------------------------', 1);
    for (const ch of chars) {
      const name = (ch.name ?? '').padEnd(25).slice(0, 25);
      const score = String(ch.scoreVal ?? 0).padStart(8);
      const alive = (ch.alive ? 'Alive' : 'Dead').padEnd(6);
      const cls = (ch.charClass ?? '').padEnd(10).slice(0, 10);
      io.println(` ${name} ${score} ${alive} ${cls}`, 1);
    }
    io.cr();
    await io.pausePrompt();
    session.setNext(new ArenaPromptState());
  }
}

export class ArenaChallengeWonState {
  async enter(session) {
    const c = session.character;
    const { io, monster } = session;
    io.cr();
    io.println(`You have Defeated ${monster.name}!!`, 3);
    const coinReward = randBetween(1, (monster.coins ?? 0) + 1);
    const xpReward = monster.experience ?? 0;
    c.totalExperience += xpReward;
    c.spendingExperience += xpReward;
    c.coinsHand += coinReward;
    c.fightsWon += 1;
    io.cr();
    io.println(`You receive ${xpReward} Experience points.`, 5);
    io.println(`You receive ${coinReward} Coins.`, 5);
    io.cr();

    // Damage the loser — kill them and write them mail.
    const loser = await findCharacterByName(session.combatName);
    if (loser) {
      loser.alive = false;
      loser.coinsHand = Math.max(0, loser.coinsHand - coinReward);
      loser.totalExperience -= 20;
      loser.spendingExperience -= 20;
      if (loser.totalExperience < 0) loser.totalExperience = 0;
      if (loser.spendingExperience < 0) loser.spendingExperience = 0;
      await saveCharacter(loser);
      await writeNews(loser.bbsName, `You have been slaughtered by ${c.name}!`);
    }
    await saveCharacter(c);
    await io.pausePrompt('-=Press A Key=-');
    c.location = WhereType.TheArenaMenu;
    session.setNext(new ArenaState());
  }
}

export class ArenaChallengeLostState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println("You're outta life buddy...", 6);
    io.cr();
    const winner = await findCharacterByName(session.combatName);
    if (winner) {
      const coinsTaken = randBetween(1, Math.floor(c.coinsHand / 2) + 1);
      const xpGained = Math.floor(c.totalExperience * 0.3);
      winner.coinsHand += coinsTaken;
      winner.totalExperience += xpGained;
      winner.spendingExperience += xpGained;
      winner.fightsWon += 1;
      winner.totalFights += 1;
      await saveCharacter(winner);
      await writeNews(winner.bbsName, `You have slaughtered ${c.name}!`);
      c.coinsHand -= coinsTaken;
    }
    c.alive = false;
    c.totalFights += 1;
    await saveCharacter(c);
    const { DeadState } = await import('./dead.js');
    session.setNext(new DeadState());
  }
}
