// Server-backed store. Mirrors the surface of the old local.js but every
// mutation roundtrips to /api/* (Pages Functions + D1). localStorage is no
// longer used — the world is shared across all players who log in.

let _currentUser = null;

// ---- auth ---------------------------------------------------------------

// Returns the BBS name of the currently logged-in player, or null. Uses an
// in-memory cache after the first /api/me probe.
export async function getActiveUser() {
  if (_currentUser !== null) return _currentUser || null;
  const res = await fetch('/api/me');
  if (!res.ok) { _currentUser = ''; return null; }
  const me = await res.json();
  _currentUser = me.bbsName || '';
  return _currentUser || null;
}

// POST /api/login. Auto-registers if the BBS name is new (the supplied
// password becomes the lock). Throws on bad credentials.
export async function login(bbsName, password) {
  const res = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bbsName, password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `login failed (${res.status})`);
  }
  const data = await res.json();
  _currentUser = data.bbsName;
  return data; // { bbsName, registered, hasCharacter }
}

export async function clearActiveUser() {
  await fetch('/api/logout', { method: 'POST' });
  _currentUser = '';
}

// ---- characters --------------------------------------------------------

export async function findCharacterByBBSName(bbsName) {
  const res = await fetch(`/api/character/${encodeURIComponent(bbsName)}`);
  if (!res.ok) return null;
  return res.json();
}

export async function saveCharacter(char) {
  const res = await fetch('/api/character', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(char),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `saveCharacter failed (${res.status})`);
  }
}

export async function findCharacterByName(name) {
  const res = await fetch(`/api/by-name/${encodeURIComponent(name)}`);
  if (!res.ok) return null;
  return res.json();
}

export async function listCharacters() {
  const res = await fetch('/api/characters');
  if (!res.ok) return [];
  return res.json();
}

// ---- inn / arena -------------------------------------------------------

function defaultInn() {
  return {
    rooms: Array.from({ length: 10 }, () => ({ who: '', daysLeft: 0, lock: 0 })),
    owner: '',
    curRate: 5,
    safe: 0,
    open: true,
  };
}

export async function loadInn() {
  const res = await fetch('/api/inn');
  if (!res.ok) return defaultInn();
  const inn = await res.json();
  return inn ?? defaultInn();
}

export async function saveInn(inn) {
  await fetch('/api/inn', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(inn),
  });
}

export async function loadGladiatorFights() {
  const res = await fetch('/api/fights');
  if (!res.ok) return [];
  return res.json();
}

export async function saveGladiatorFight(fight) {
  const fights = await loadGladiatorFights();
  fights.push(fight);
  await saveAllGladiatorFights(fights);
}

export async function saveAllGladiatorFights(fights) {
  await fetch('/api/fights', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(fights),
  });
}

export async function loadBets() {
  const res = await fetch('/api/bets');
  if (!res.ok) return [];
  return res.json();
}

export async function saveBet(bet) {
  const bets = await loadBets();
  bets.push(bet);
  await saveAllBets(bets);
}

export async function saveAllBets(bets) {
  await fetch('/api/bets', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(bets),
  });
}

// ---- news --------------------------------------------------------------

// Atomic fetch-and-clear of the current user's news. The bbsName arg is
// kept for parity with local.js but ignored — the server reads identity
// from the session cookie. Mailbox-of-other-player reads were never used
// by the codebase.
export async function readNews(_bbsName) {
  const res = await fetch('/api/news');
  if (!res.ok) return '';
  const j = await res.json();
  return j.text ?? '';
}

// No-op: readNews already deletes the rows server-side. Kept so existing
// callers (enter_slycrel) compile without edits.
export async function clearNews(_bbsName) {}

export async function writeNews(bbsName, text) {
  await fetch('/api/news/send', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ to: bbsName, body: text }),
  });
}

// ---- static data files (same as local.js) ------------------------------

let _weaponsCache = null, _armorCache = null;
const _terrainCache = {}, _monsterCache = {};

export async function loadWeapons() {
  if (_weaponsCache) return _weaponsCache;
  const res = await fetch('/data/weapons.json');
  _weaponsCache = await res.json();
  return _weaponsCache;
}

export async function loadArmor() {
  if (_armorCache) return _armorCache;
  const res = await fetch('/data/armor.json');
  _armorCache = await res.json();
  return _armorCache;
}

export async function loadTerrain(mapName) {
  if (_terrainCache[mapName]) return _terrainCache[mapName];
  const res = await fetch(`/data/terrain/${mapName}.json`);
  if (!res.ok) return null;
  _terrainCache[mapName] = await res.json();
  return _terrainCache[mapName];
}

export async function loadMonsters(region) {
  if (_monsterCache[region]) return _monsterCache[region];
  const res = await fetch(`/data/monsters/${region}.json`);
  if (!res.ok) return [];
  _monsterCache[region] = await res.json();
  return _monsterCache[region];
}

export async function getRandomMonster(region, playerLevel) {
  const all = await loadMonsters(region);
  if (!all.length) return null;
  const candidates = all.filter(m => m.level <= playerLevel + 2 && m.level >= playerLevel - 2);
  const pool = candidates.length ? candidates : all;
  return { ...pool[Math.floor(Math.random() * pool.length)] };
}
