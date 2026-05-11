// Browser-side persistence. Mirrors the surface of internal/store/Store
// just enough to satisfy the states we've ported. Keyed by BBS name so the
// future option-2 (server-authoritative) port can swap the backing without
// touching callers.

const CHAR_KEY = (bbsName) => `slycrel.character.${bbsName}`;
const ACTIVE_USER_KEY = 'slycrel.active_user';

export function findCharacterByBBSName(bbsName) {
  const raw = localStorage.getItem(CHAR_KEY(bbsName));
  return raw ? JSON.parse(raw) : null;
}

export function saveCharacter(char) {
  localStorage.setItem(CHAR_KEY(char.bbsName), JSON.stringify(char));
}

export function getActiveUser() {
  return localStorage.getItem(ACTIVE_USER_KEY);
}

export function setActiveUser(bbsName) {
  localStorage.setItem(ACTIVE_USER_KEY, bbsName);
}

export function clearActiveUser() {
  localStorage.removeItem(ACTIVE_USER_KEY);
}

// Read-only catalogs shipped under /data. Cached after first load.
let _weaponsCache = null;
let _armorCache = null;

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

const _terrainCache = {};
export async function loadTerrain(mapName) {
  if (_terrainCache[mapName]) return _terrainCache[mapName];
  const res = await fetch(`/data/terrain/${mapName}.json`);
  if (!res.ok) return null;
  _terrainCache[mapName] = await res.json();
  return _terrainCache[mapName];
}

const _monsterCache = {};
export async function loadMonsters(region) {
  if (_monsterCache[region]) return _monsterCache[region];
  const res = await fetch(`/data/monsters/${region}.json`);
  if (!res.ok) return [];
  _monsterCache[region] = await res.json();
  return _monsterCache[region];
}

// Mirrors JSONStore.GetRandomMonster: prefer monsters within ±2 of the
// player's level, fall back to all monsters if none match.
export async function getRandomMonster(region, playerLevel) {
  const all = await loadMonsters(region);
  if (!all.length) return null;
  const candidates = all.filter(m => m.level <= playerLevel + 2 && m.level >= playerLevel - 2);
  const pool = candidates.length ? candidates : all;
  return { ...pool[Math.floor(Math.random() * pool.length)] };
}

// Inn state — shared across all players in Go, single-keyed here.
const INN_KEY = 'slycrel.inn';

function defaultInn() {
  return {
    rooms: Array.from({ length: 10 }, () => ({ who: '', daysLeft: 0, lock: 0 })),
    owner: '',
    curRate: 5,
    safe: 0,
    open: true,
  };
}

export function loadInn() {
  const raw = localStorage.getItem(INN_KEY);
  if (!raw) return defaultInn();
  try { return JSON.parse(raw); } catch { return defaultInn(); }
}

export function saveInn(inn) {
  localStorage.setItem(INN_KEY, JSON.stringify(inn));
}

// Find a character by their in-game name (not BBS login). Walks all saved
// characters — slow at scale, but fine for a localStorage-backed game.
export function findCharacterByName(name) {
  const needle = name.toLowerCase();
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (!key || !key.startsWith('slycrel.character.')) continue;
    try {
      const c = JSON.parse(localStorage.getItem(key));
      if (c.name?.toLowerCase() === needle) return c;
    } catch {}
  }
  return null;
}

// Arena gladiator fights + bets — shared mutable state.
const FIGHTS_KEY = 'slycrel.gladiator_fights';
const BETS_KEY = 'slycrel.gladiator_bets';

export function loadGladiatorFights() {
  const raw = localStorage.getItem(FIGHTS_KEY);
  if (!raw) return [];
  try { return JSON.parse(raw); } catch { return []; }
}

export function saveGladiatorFight(fight) {
  const fights = loadGladiatorFights();
  fights.push(fight);
  localStorage.setItem(FIGHTS_KEY, JSON.stringify(fights));
}

export function saveAllGladiatorFights(fights) {
  localStorage.setItem(FIGHTS_KEY, JSON.stringify(fights));
}

export function loadBets() {
  const raw = localStorage.getItem(BETS_KEY);
  if (!raw) return [];
  try { return JSON.parse(raw); } catch { return []; }
}

export function saveBet(bet) {
  const bets = loadBets();
  bets.push(bet);
  localStorage.setItem(BETS_KEY, JSON.stringify(bets));
}

export function saveAllBets(bets) {
  localStorage.setItem(BETS_KEY, JSON.stringify(bets));
}

// Per-player news/mail. Keyed by BBS name so the message survives logins.
const NEWS_KEY = (bbsName) => `slycrel.news.${bbsName}`;

export function readNews(bbsName) {
  return localStorage.getItem(NEWS_KEY(bbsName)) ?? '';
}

export function writeNews(bbsName, text) {
  const prev = readNews(bbsName);
  const next = prev ? prev + '\n' + text : text;
  localStorage.setItem(NEWS_KEY(bbsName), next);
}

export function clearNews(bbsName) {
  localStorage.removeItem(NEWS_KEY(bbsName));
}

// Walk localStorage and return every saved character. For a browser-only
// build that's usually just the active player, but Tavern's View Guilds
// still wants to enumerate them. Mirrors store.ListCharacters in Go.
export function listCharacters() {
  const out = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (!key || !key.startsWith('slycrel.character.')) continue;
    try { out.push(JSON.parse(localStorage.getItem(key))); } catch {}
  }
  return out;
}
