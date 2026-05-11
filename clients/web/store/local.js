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
