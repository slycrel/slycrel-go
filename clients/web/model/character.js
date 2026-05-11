import { CharClass, Color, WhereType } from './enums.js';

// Starting-value constants from the Go side (internal/model/constants.go-ish —
// they live alongside the enums there). Mirrored as plain consts.
export const BASE_EXP = 0;
export const STARTING_COINS_HAND = 100;
export const STARTING_COINS_BANK = 0;
export const MAX_EXPLORATION = 15;
export const MAX_SPARS = 4;

// Build an empty character keyed by BBS name. rollChar() fills in stats.
export function newBlankCharacter(bbsName, name) {
  return {
    fileIndex: 0,
    name: name ?? bbsName,
    bbsName,
    scoreVal: 0,
    lastOn: 0,
    alive: true,
    location: WhereType.TheTown,
    gender: true,
    charClass: CharClass.Fighter,
    totalExperience: 0,
    spendingExperience: 0,
    thiefLvl: 0, mageLvl: 0, fighterLvl: 0,
    hitPoints: 0, maxHP: 0,
    coinsHand: 0, coinsBank: 0,
    speed: 0, strength: 0, dexterity: 0,
    psyche: 0, maxPsyche: 0,
    fame: 0, honor: 0, faith: 0,
    favColor: Color.Blue,
    flirt1: 0, flirt2: 0,
    ammo: 0, movement: 0,
    exploration: 0, spars: 0,
    totalFights: 0, fightsWon: 0,
    weapons: [emptyWeapon(), emptyWeapon(), emptyWeapon()],
    armor: [emptyArmor(), emptyArmor()],
  };
}

export function emptyWeapon() {
  return { name: '', strike: 0, range: 0, actionStr: '' };
}

export function emptyArmor() {
  return { name: '', defense: 0 };
}

export function isWeaponEmpty(w) {
  return !w || !w.name;
}

export function isArmorEmpty(a) {
  return !a || !a.name;
}

export function genderString(c) {
  return c.gender ? 'Male' : 'Female';
}
