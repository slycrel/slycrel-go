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
