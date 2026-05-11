import { GRID_ROWS, GRID_COLS, TerrainCell, WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import {
  TERRAIN_DISPLAY, numMovePointsAt, canMoveAnywhere, findRoute,
} from '../../mechanics/terrain.js';
import { loadTerrain, saveCharacter } from '../../store/local.js';
import {
  SetupTextCombatState, UserKilledState, UserVictoriousState,
} from './combat_text.js';

const LOG_SIZE = 5;

// Pure HTML rendering — we have the full RGB palette, so just emit a
// styled span per cell into a <pre>. One sweep, set innerHTML once.
function renderGridScene(session) {
  const { io, character: c, monster, terrain } = session;
  if (!terrain) return;
  const lines = [];
  for (let r = 0; r < GRID_ROWS; r++) {
    let row = '';
    for (let col = 0; col < GRID_COLS; col++) {
      let ch, fg, bg;
      if (r === session.userR && col === session.userC) {
        ch = '@'; fg = '#ff0'; bg = '#000';
      } else if (r === session.monsR && col === session.monsC) {
        ch = 'M'; fg = '#f55'; bg = '#000';
      } else {
        ({ ch, fg, bg } = TERRAIN_DISPLAY[terrain.cells[r][col]] ?? TERRAIN_DISPLAY[0]);
      }
      row += `<span style="color:${fg};background:${bg}">${ch === ' ' ? '&nbsp;' : ch}</span>`;
    }
    lines.push(row);
  }
  const grid = lines.join('\n');

  const hud =
    `<span style="color:#cdcd00">${c.name}</span>  ` +
    `<span style="color:#0c0">HP:${c.hitPoints}/${c.maxHP}</span>  ` +
    `<span style="color:#fc0">Move:${c.movement}</span>  ` +
    `<span style="color:#0cc">Wpn:${c.weapons[0].name || 'fists'}` +
    (c.weapons[1].name ? ` / ${c.weapons[1].name}` : '') + `</span>\n` +
    `<span style="color:#f55">${monster.name}: ${monster.hitPoints} HP</span>`;

  const log = activeGridLog(session)
    .map(l => `<span style="color:#cdcd00">${escapeHtml(l)}</span>`)
    .join('\n');

  io.clear();
  const pre = document.createElement('pre');
  pre.style.lineHeight = '1.0';
  pre.innerHTML = grid + '\n\n' + hud + (log ? '\n\n' + log : '');
  io.root.appendChild(pre);
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, ch => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  }[ch]));
}

function activeGridLog(session) {
  const buf = session.gridStatusMsgs;
  const start = session.textOutln % buf.length;
  const out = [];
  for (let i = 0; i < buf.length; i++) {
    const msg = buf[(start + i) % buf.length];
    if (msg) out.push(msg);
  }
  return out;
}

function gridTextOut(session, msg) {
  session.gridStatusMsgs[session.textOutln % session.gridStatusMsgs.length] = msg;
  session.textOutln += 1;
  renderGridScene(session);
}

function moveGridPlayer(session, dr, dc) {
  const c = session.character;
  let nr = session.userR + dr, nc = session.userC + dc;
  if (nr < 0) nr = 0;
  if (nr >= GRID_ROWS) nr = GRID_ROWS - 1;
  if (nc < 0) nc = 0;
  if (nc >= GRID_COLS) nc = GRID_COLS - 1;
  const cost = numMovePointsAt(nr, nc, session.terrain);
  if (cost >= 998) {
    if (session.terrain.cells[nr][nc] === TerrainCell.Water) {
      gridTextOut(session, 'You fall into the water and drown!');
      c.alive = false;
      c.hitPoints = 0;
      return;
    }
    gridTextOut(session, "You can't move there!");
    return;
  }
  if (cost > c.movement) {
    gridTextOut(session, 'Not enough movement points!');
    c.movement -= 1;
    if (c.movement < 0) c.movement = 0;
    return;
  }
  session.userR = nr;
  session.userC = nc;
  c.movement -= cost;
  renderGridScene(session);
}

function doGridShoot(session) {
  const c = session.character;
  const m = session.monster;
  c.movement -= 3;
  const drv = session.monsR - session.userR;
  const dcv = session.monsC - session.userC;
  const dist = Math.sqrt(drv * drv + dcv * dcv);
  let weapRange = c.weapons[1].range ?? 0;
  if (weapRange <= 0) weapRange = c.weapons[0].range ?? 0;
  if (weapRange <= 0) {
    gridTextOut(session, 'You have no ranged weapon!');
    return;
  }
  if (Math.floor(dist) > weapRange) {
    gridTextOut(session, 'Your Shot Fell Short!');
    return;
  }
  const hitChance = c.dexterity * 3 - Math.floor(dist) * 2;
  if (randBetween(1, 100) > hitChance) {
    const misses = [
      'Your Shot Fell Short!',
      'Your Shot Bounces off a Boulder!',
      "Your Aim isn't very good, it hit a TREE!",
      'Aww.. Too Bad, Hit a Tree...',
      'Argh! Take some Lessons!',
    ];
    gridTextOut(session, misses[randBetween(0, misses.length - 1)]);
    return;
  }
  let strike = c.weapons[1].strike ?? 0;
  if (strike <= 0) strike = c.weapons[0].strike ?? 0;
  let dmg = randBetween(Math.floor(strike / 2), strike) +
            randBetween(0, Math.floor(c.strength / 3));
  dmg -= Math.floor((m.defense ?? 0) / 2);
  if (dmg < 1) dmg = 1;
  m.hitPoints -= dmg;
  gridTextOut(session, `You hit ${m.name} for ${dmg} Damage!`);
}

function doGridMonsterShoot(session) {
  const c = session.character;
  const m = session.monster;
  if ((m.rangeOffense ?? 0) <= 0) return;
  const drv = session.userR - session.monsR;
  const dcv = session.userC - session.monsC;
  const dist = Math.sqrt(drv * drv + dcv * dcv);
  if (Math.floor(dist) > (m.range ?? 0)) return;
  if (randBetween(1, 100) > 60) {
    gridTextOut(session, `The ${m.name} shoots and misses!`);
    return;
  }
  let dmg = randBetween(Math.floor(m.rangeOffense / 2), m.rangeOffense) -
            Math.floor((c.armor[0]?.defense ?? 0) / 2);
  if (dmg < 1) dmg = 1;
  c.hitPoints -= dmg;
  gridTextOut(session, `The ${m.name} hits you for ${dmg} ranged damage!`);
}

// GridCombatSetupState — picks a terrain map for the region, places both
// combatants on passable cells, falls through to text combat if no map.
export class GridCombatSetupState {
  async enter(session) {
    const region = session.combatRegion || 'forest';
    const n = randBetween(1, 3);
    const mapName = `${region}-${n}`;
    const terrain = await loadTerrain(mapName);
    if (!terrain) {
      session.setNext(new SetupTextCombatState());
      return;
    }
    session.terrain = terrain;

    // Place player.
    for (let i = 0; i < 100; i++) {
      session.userR = randBetween(0, GRID_ROWS - 1);
      session.userC = randBetween(0, GRID_COLS - 1);
      if (numMovePointsAt(session.userR, session.userC, terrain) < 998) break;
    }
    // Place monster — passable cell, not on the player.
    for (let i = 0; i < 100; i++) {
      session.monsR = randBetween(0, GRID_ROWS - 1);
      session.monsC = randBetween(0, GRID_COLS - 1);
      const cost = numMovePointsAt(session.monsR, session.monsC, terrain);
      if (cost < 998 && !(session.monsR === session.userR && session.monsC === session.userC)) break;
    }

    session.textOutln = 0;
    session.gridStatusMsgs = ['', '', '', '', ''];
    gridTextOut(session, `You Encounter a ${session.monster.name}`);
    session.setNext(new GridCombatPromptState());
  }
}

export class GridCombatPromptState {
  async enter(session) {
    const { io, character: c, monster } = session;
    if (c.hitPoints <= 0 || !c.alive) {
      session.setNext(new UserKilledState());
      return;
    }
    if (monster.hitPoints <= 0) {
      io.clear();
      session.setNext(new UserVictoriousState());
      return;
    }
    if (!canMoveAnywhere(session.userR, session.userC, c.movement, session.terrain)) {
      session.setNext(new GridMonsterMoveState());
      return;
    }

    renderGridScene(session);
    const choice = await io.lettersPrompt('IJKL move  A attack  F fire  R run  P pass', 'AIJKLFRPS*');

    switch (choice) {
      case 'I': moveGridPlayer(session, -1, 0); break;
      case 'K': moveGridPlayer(session, 1, 0); break;
      case 'J': moveGridPlayer(session, 0, -1); break;
      case 'L': moveGridPlayer(session, 0, 1); break;
      case 'A':
        if (c.movement < 3) {
          gridTextOut(session, "Not 'nuff Movement Pts. Left!");
        } else {
          c.movement -= 3;
          if (randBetween(1, 50) < 8) {
            session.setNext(new SetupTextCombatState());
            return;
          }
          gridTextOut(session, 'You Attempt to Attack the Creature, & you fail.');
        }
        break;
      case 'F':
        if (c.movement < 3) {
          gridTextOut(session, "Not 'nuff Movement Pts. Left!");
        } else {
          doGridShoot(session);
        }
        break;
      case 'R':
        if (c.location === WhereType.TheArenaCombat) {
          c.location = WhereType.TheArenaMenu;
          const { ArenaState } = await import('./arena.js');
          session.setNext(new ArenaState());
        } else {
          c.location = WhereType.TheWildernessMenu;
          const { WildernessState } = await import('./wilderness.js');
          session.setNext(new WildernessState());
        }
        return;
      case 'P':
        gridTextOut(session, 'You pass your remaining moves');
        session.setNext(new GridMonsterMoveState());
        return;
      case 'S':
        gridTextOut(session, 'You feel special while passing your moves');
        session.setNext(new GridMonsterMoveState());
        return;
      case '*':
        renderGridScene(session);
        break;
    }

    if (session.userR === session.monsR && session.userC === session.monsC) {
      session.setNext(new SetupTextCombatState());
      return;
    }
    session.setNext(new GridCombatPromptState());
  }
}

export class GridMonsterMoveState {
  async enter(session) {
    const { character: c, monster, terrain } = session;
    const route = findRoute(session.monsR, session.monsC, session.userR, session.userC, terrain);
    if (route) {
      let moveLeft = monster.movement ?? 0;
      for (const d of route) {
        let nr = session.monsR, nc = session.monsC;
        if (d === 'U') nr -= 1;
        else if (d === 'D') nr += 1;
        else if (d === 'L') nc -= 1;
        else if (d === 'R') nc += 1;
        if (nr < 0 || nr >= GRID_ROWS || nc < 0 || nc >= GRID_COLS) continue;
        const cost = numMovePointsAt(nr, nc, terrain);
        if (cost > moveLeft) break;
        session.monsR = nr;
        session.monsC = nc;
        moveLeft -= cost;
        if (randBetween(1, 5) === 1 && moveLeft >= 3) {
          doGridMonsterShoot(session);
          moveLeft -= 3;
        }
        if (moveLeft < 1) break;
      }
      renderGridScene(session);
    }
    const spd = Math.max(c.speed, 5);
    const spread = Math.floor(spd / 5);
    c.movement = randBetween(spd - spread, spd + spread);
    saveCharacter(c);
    if (session.userR === session.monsR && session.userC === session.monsC) {
      session.setNext(new SetupTextCombatState());
      return;
    }
    session.setNext(new GridCombatPromptState());
  }
}
