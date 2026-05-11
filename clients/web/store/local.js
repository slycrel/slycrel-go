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
