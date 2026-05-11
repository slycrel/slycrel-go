import { CombatType, HitType, WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import {
  rollInitiative, doDamageUser, doDamageMonster, randMonsterAction,
  calcAdvanceMoves, applyUserAttackModifier, applyMonsterAttackModifier,
  getDisposition,
} from '../../mechanics/combat.js';
import { saveCharacter, getRandomMonster } from '../../store/local.js';
import { level } from '../../model/character.js';
import { DeadState } from './dead.js';

// Port of internal/game/states/combat_text.go.
// Text combat is the menu-driven A/B/P/D/S/R loop — no grid, no tiles.

const INIT_TEXTS_PLAYER_FIRST = [
  'You close with the vile %s, and snarl in defiance.\nThe %s quakes in fear as you ready your %s and square off...',
  'The terrifying %s quickly closes as you begin laughing.\nThis is going to be as easy as flirting with Corenne...',
  'A strange feeling comes over you as you begin the melee.\nA curse of rage is thrown your way from the %s in a guttural tongue...',
  'The %s leaps at you, inhuman appendages flailing.\nYou bring up your arms in defense as %s crashes into you. The battle begins!',
  'You ready your %s and prepare for the oncoming fray.\nThe surrounding terrain goes silent as you face off with the %s.',
  'The %s looks surprised that you do not quake in fear, and\nstarts waving its %s menacingly, attempting to ward you off.\nYou gleefully close for the kill!',
  'As you close with the %s, you cannot help but\ncompare it to your sixth grade teacher, Grizelda. As\nthe fight begins, you resolve to rid the world of both hideous beings...',
];

const INIT_TEXTS_MONSTER_FIRST = [
  'You cannot flee. The %s slashes at your back,\ncutting a painful gash in your left shoulder. You turn,\nand ready your %s, convinced that you are about to die...',
  'You bring up your %s, assured that you can\ndefeat the %s. As silence descends on the surrounding\narea, you start to wonder where the %s has gone to.\nThe sudden slice in your back whips you around, the fight already started...',
  'You bring up your %s, assured that you can\ndefeat the %s. As silence descends, you begin to\nwonder where that %s has gone to. The sudden\nslice in your back flips you around, the fiend already\npreparing for another strike...',
];

// printf-style %s replacement (positional, left to right).
function fmt(template, ...args) {
  let i = 0;
  return template.replace(/%s/g, () => args[i++] ?? '');
}

function printLines(io, text, color) {
  for (const line of text.split('\n')) io.println(line, color);
}

// Brief pause so a flurry of state-machine transitions doesn't paint
// the monster's response and the next prompt onto the screen in the
// same frame. ~400ms is enough to read a one-line action.
const COMBAT_PAUSE_MS = 400;
function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

// SetupCombatState — port of wilderness.go SetupCombatState. Decrements
// exploration, loads a random region monster, then routes to text combat.
// The Go version goes to grid_combat first; we skip that until phase 5.
export class SetupCombatState {
  async enter(session) {
    const { io, character: c } = session;
    c.totalFights += 1;
    if (c.exploration <= 0) {
      io.println("You just can't find any more monsters today... Try again tomorrow.", 3);
      io.cr();
      await io.pausePrompt('-=Press A Key=-');
      const { WildernessState } = await import('./wilderness.js');
      session.setNext(new WildernessState());
      return;
    }
    c.exploration -= 1;
    const monster = await getRandomMonster(session.combatRegion, level(c));
    if (!monster) {
      io.println('The wilderness is eerily quiet... no monsters found.', 3);
      await io.pausePrompt();
      const { WildernessState } = await import('./wilderness.js');
      session.setNext(new WildernessState());
      return;
    }
    session.monster = monster;
    session.monHitPoints = monster.hitPoints;
    const spread = Math.floor(Math.max(c.speed, 5) / 5);
    c.movement = randBetween(Math.max(c.speed, 5) - spread, Math.max(c.speed, 5) + spread);
    io.clear();
    io.cr();
    io.println(`You encounter a ${monster.name}!`, 6);
    io.cr();
    saveCharacter(c);
    // Hand off to the grid combat setup — it'll fall back to text combat
    // if no terrain map exists for the region.
    const { GridCombatSetupState } = await import('./combat_grid.js');
    session.setNext(new GridCombatSetupState());
  }
}

export class SetupTextCombatState {
  async enter(session) {
    const { io, character: c, monster } = session;
    io.clear();
    io.cr();

    session.userAttackLoop = Math.max(1, Math.floor(c.speed / 10));
    session.monAttackLoop = Math.max(1, Math.floor((monster.movement ?? 0) / 10));
    session.textCombatLoop = 15;
    session.hit = HitType.None;
    session.userMove = CombatType.Attack;
    session.monMove = CombatType.Attack;
    session.lastDisposition = 0;

    const gotInit = rollInitiative(c.movement, monster.movement ?? 0);
    showInitiativeText(session, gotInit);
    session.setNext(new TextCombatLoopState());
  }
}

function showInitiativeText(session, playerFirst) {
  const { io, character: c, monster } = session;
  const monName = monster.name;
  const weapName = c.weapons[0].name || 'fists';
  const monWeap = monster.attackStr1 ?? '';

  if (playerFirst) {
    const idx = Math.floor(Math.random() * INIT_TEXTS_PLAYER_FIRST.length);
    const t = INIT_TEXTS_PLAYER_FIRST[idx];
    let text;
    switch (idx) {
      case 0: text = fmt(t, monName, monName, weapName); break;
      case 1: text = fmt(t, monName); break;
      case 2: text = fmt(t, monName); break;
      case 3: text = fmt(t, monName, monName); break;
      case 4: text = fmt(t, weapName, monName); break;
      case 5: text = fmt(t, monName, monWeap); break;
      case 6: text = fmt(t, monName); break;
    }
    printLines(io, text, 0);
  } else {
    const idx = Math.floor(Math.random() * INIT_TEXTS_MONSTER_FIRST.length);
    const t = INIT_TEXTS_MONSTER_FIRST[idx];
    let text;
    switch (idx) {
      case 0: text = fmt(t, monName, weapName); break;
      case 1: text = fmt(t, weapName, monName, monName); break;
      case 2: text = fmt(t, weapName, monName, monName); break;
    }
    printLines(io, text, 0);
    const dmg = doDamageMonster(c, monster);
    if (dmg > 0) {
      c.hitPoints -= dmg;
      io.cr();
      io.println(`You are hit for ${dmg} damage.`, 6);
    }
  }
  io.cr();
}

export class TextCombatLoopState {
  async enter(session) {
    const { character: c, monster } = session;
    if (c.hitPoints <= 0) { session.setNext(new UserKilledState()); return; }
    if (monster.hitPoints <= 0) { session.setNext(new UserVictoriousState()); return; }
    if (session.hit === HitType.UserRan) { session.setNext(new UserEscapesState()); return; }
    if (session.hit === HitType.OpponentRan) { session.setNext(new MonsterRunsState()); return; }

    if (session.userAttackLoop >= session.textCombatLoop && session.hit === HitType.None) {
      session.setNext(new UserAttackStageState());
      return;
    }
    if (session.monAttackLoop >= session.textCombatLoop &&
        session.hit !== HitType.Opponent &&
        session.hit !== HitType.OpponentOnly &&
        session.hit !== HitType.Both &&
        session.hit !== HitType.Neither &&
        session.hit !== HitType.UserOnly &&
        session.hit !== HitType.OpponentRan) {
      session.setNext(new OpponentAttackStageState());
      return;
    }
    session.textCombatLoop -= 1;
    session.hit = HitType.None;
    if (session.textCombatLoop <= 0) session.textCombatLoop = 27;
    session.setNext(new TextCombatLoopState());
  }
}

export class UserAttackStageState {
  async enter(session) {
    const { io, character: c, monster } = session;
    showDispositionText(session);
    io.cr();
    const choice = await io.lettersPrompt('A]ttack, B]lock, P]arry, D]odge, S]pecial, R]un --', 'ABPDSR');
    const advMoves = calcAdvanceMoves(session.userAttackLoop, session.monAttackLoop, session.textCombatLoop);

    switch (choice) {
      case 'A': {
        session.userMove = CombatType.Attack;
        io.println('You attack the sucker!', 1);
        io.cr();
        let dmg = doDamageUser(c, monster);
        dmg = applyUserAttackModifier(dmg, session.monMove);
        if (session.monMove === CombatType.Dodge) {
          if (randBetween(1, c.dexterity) < Math.floor((monster.movement ?? 0) / 3)) dmg = 0;
          if (dmg === 0) { io.println('The monster laughs as it dodges your attack.', 5); io.cr(); }
        }
        let total = dmg;
        if (total > 0) {
          io.println(`You hit the monster for ${total} damage.`, 2);
          session.hit = HitType.TheUser;
        } else {
          io.cr();
          io.println('You miss!', 6);
          io.cr();
        }
        for (let i = 0; i < advMoves; i++) {
          const extra = doDamageUser(c, monster);
          total += extra;
          io.println(`Again, You hit the vile fiend for ${extra} Damage!`, 1);
          session.hit = HitType.TheUser;
        }
        monster.hitPoints -= total;
        session.hit = HitType.TheUser;
        break;
      }
      case 'B':
        session.userMove = CombatType.Block;
        session.hit = HitType.UserMiss;
        io.println('You raise your guard, ready to block.', 3);
        break;
      case 'P': {
        session.userMove = CombatType.Parry;
        session.hit = HitType.TheUser;
        const dmg = Math.floor(doDamageUser(c, monster) / 3);
        if (dmg > 0) io.println(`You parry, slicing the monster for ${dmg} damage.`, 1);
        else io.println('You raise your weapon, ready to parry.', 3);
        io.cr();
        monster.hitPoints -= dmg;
        break;
      }
      case 'D':
        session.userMove = CombatType.Dodge;
        session.hit = HitType.UserMiss;
        io.println('You prepare to dodge...', 3);
        break;
      case 'S':
        session.userMove = CombatType.Special;
        session.hit = HitType.UserMiss;
        io.cr();
        io.println('Okay, you sit there and feel special.', 1);
        break;
      case 'R':
        io.cr();
        io.println('You attempt escape...', 2);
        io.cr();
        session.hit = HitType.UserRan;
        session.userMove = CombatType.Special;
        break;
    }
    io.cr();
    session.setNext(new TextCombatLoopState());
  }
}

export class OpponentAttackStageState {
  async enter(session) {
    const { io, character: c, monster } = session;
    const advMoves = calcAdvanceMoves(session.monAttackLoop, session.userAttackLoop, session.textCombatLoop);
    const { action, ran } = randMonsterAction(monster.hitPoints, session.monHitPoints);
    if (ran) {
      session.hit = HitType.OpponentRan;
      session.setNext(new TextCombatLoopState());
      return;
    }

    switch (action) {
      case CombatType.Attack: {
        session.monMove = CombatType.Attack;
        let dmg = doDamageMonster(c, monster);
        dmg = applyMonsterAttackModifier(dmg, session.userMove, c.dexterity, monster.movement ?? 0);
        if (session.userMove === CombatType.Dodge && dmg === 0) {
          io.println('You Dodge successfully!', 5);
          io.cr();
        }
        if (session.userMove === CombatType.Special && dmg > 0) {
          io.println('You feel VERY special now.', 2);
          io.cr();
        }
        let total = dmg;
        if (total > 0) {
          io.cr();
          io.println('The monster attacks you!', 3);
          io.cr();
          io.println(`You are hit for ${total} damage. (Feel better yet?)`, 1);
          io.cr();
          for (let i = 0; i < advMoves; i++) {
            let extra = doDamageMonster(c, monster);
            extra = applyMonsterAttackModifier(extra, session.userMove, c.dexterity, monster.movement ?? 0);
            total += extra;
            io.println(`You are hit again for ${extra} damage!`, 6);
          }
          c.hitPoints -= total;
          if (session.hit === HitType.TheUser) session.hit = HitType.Both;
          else if (session.hit === HitType.UserMiss || session.hit === HitType.None) session.hit = HitType.OpponentOnly;
        } else {
          io.println('The monster misses you!', 3);
          if (session.hit === HitType.TheUser) session.hit = HitType.UserOnly;
          else if (session.hit === HitType.UserMiss || session.hit === HitType.None) session.hit = HitType.Neither;
        }
        break;
      }
      case CombatType.Block:
        io.println('The monster keeps its distance, circling warily.', 1);
        io.cr();
        session.monMove = CombatType.Block;
        session.hit = (session.hit === HitType.TheUser) ? HitType.UserOnly : HitType.Neither;
        break;
      case CombatType.Parry:
        io.cr();
        io.println("The monster darts in and out, you aren't quite sure if it's attacking or not.", 2);
        io.cr();
        session.monMove = CombatType.Parry;
        session.hit = (session.hit === HitType.TheUser) ? HitType.UserOnly : HitType.OpponentOnly;
        break;
      case CombatType.Dodge:
        io.cr();
        io.println('The monster backs away, hissing at you, looking like a coiled spring, ready to snap...', 3);
        io.cr();
        session.monMove = CombatType.Dodge;
        session.hit = (session.hit === HitType.TheUser) ? HitType.UserOnly : HitType.Neither;
        break;
    }
    await sleep(COMBAT_PAUSE_MS);
    session.setNext(new TextCombatLoopState());
  }
}

export class UserKilledState {
  async enter(session) {
    const { io, character: c } = session;
    if (c.location === WhereType.TheArenaCombat) {
      c.location = WhereType.TheArenaMenu;
      const { ArenaChallengeLostState } = await import('./arena.js');
      session.setNext(new ArenaChallengeLostState());
      return;
    }
    io.cr();
    io.println("Tough luck. You're dead, buddy.", 4);
    io.println("Seeya in Hero's Heaven!", 5);
    io.cr();
    c.coinsHand -= Math.floor(c.coinsHand / 3);
    c.alive = false;
    saveCharacter(c);
    session.setNext(new DeadState());
  }
}

export class UserVictoriousState {
  async enter(session) {
    const { io, character: c, monster } = session;
    if (c.location === WhereType.TheArenaCombat) {
      c.location = WhereType.TheArenaMenu;
      const { ArenaChallengeWonState } = await import('./arena.js');
      session.setNext(new ArenaChallengeWonState());
      return;
    }
    io.cr();
    io.println(`You have defeated the ${monster.name}!!`, 3);
    io.cr();
    const coinReward = Math.floor(Math.random() * ((monster.coins ?? 0) + 1));
    const xpReward = monster.experience ?? 0;
    c.totalExperience += xpReward;
    c.spendingExperience += xpReward;
    c.coinsHand += coinReward;
    c.fightsWon += 1;
    io.println(`You receive ${xpReward} Experience points.`, 5);
    io.println(`You receive ${coinReward} Coins.`, 5);
    io.cr();
    io.println(`Hit Points: ${c.hitPoints}/${c.maxHP}`, 1);
    io.cr();
    c.location = WhereType.TheWildernessMenu;
    saveCharacter(c);
    await io.pausePrompt('-=Press A Key=-');
    const { WildernessState } = await import('./wilderness.js');
    session.setNext(new WildernessState());
  }
}

export class UserEscapesState {
  async enter(session) {
    const { io, character: c } = session;
    io.cr();
    io.println('You try to escape...', 3);
    io.println('.'.repeat(40), 2);
    io.cr();
    io.println('You Succeed!', 6);
    io.cr();
    await io.pausePrompt();
    c.location = WhereType.TheWildernessMenu;
    saveCharacter(c);
    const { WildernessState } = await import('./wilderness.js');
    session.setNext(new WildernessState());
  }
}

export class MonsterRunsState {
  async enter(session) {
    const { io, character: c } = session;
    io.cr();
    io.println('<blink blink>', 1);
    io.cr();
    io.println('The monster ran away from you! What a coward.', 1);
    io.cr();
    await io.pausePrompt('-More-');
    c.location = WhereType.TheWildernessMenu;
    saveCharacter(c);
    const { WildernessState } = await import('./wilderness.js');
    session.setNext(new WildernessState());
  }
}

function showDispositionText(session) {
  const { io, character: c, monster } = session;
  const disp = getDisposition(c.hitPoints, c.maxHP, monster.hitPoints, session.monHitPoints);
  if (disp === session.lastDisposition) return;
  session.lastDisposition = disp;
  const monName = monster.name;
  const userPct = c.maxHP > 0 ? Math.round(c.hitPoints / c.maxHP * 100) : 0;
  const monPct = session.monHitPoints > 0 ? Math.round(monster.hitPoints / session.monHitPoints * 100) : 0;
  io.cr();
  switch (disp) {
    case 1: io.println(`Both you and the ${monName} are in good shape. This could go either way...`, 1); break;
    case 2: io.println(`The ${monName} looks barely scratched, while you can barely stand...`, 6); break;
    case 3: io.println(`The ${monName} is staggering! You can taste victory!`, 3); break;
    case 4: io.println(`You and the ${monName} trade blows evenly...`, 4); break;
    case 5: io.println(`You have the advantage! The ${monName} is starting to weaken.`, 3); break;
    case 6: io.println(`Things aren't looking great... the ${monName} presses its advantage.`, 6); break;
    case 7: io.println(`The ${monName} is still going strong. You need to press harder!`, 4); break;
    case 8: io.println(`The ${monName} is weakening! Keep it up!`, 3); break;
    case 9: io.println(`You and the ${monName} are both barely standing... one more hit could end it!`, 6); break;
  }
  io.println(`  [You: ${userPct}% | ${monName}: ${monPct}%]`, 1);
  io.cr();
}
