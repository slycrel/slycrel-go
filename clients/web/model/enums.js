// Enums mirroring internal/model/enums.go.
// CharClass is a string type on the Go side; mirror that here so saved
// characters round-trip through JSON without translation.

export const CharClass = Object.freeze({
  Fighter: 'Fighter',
  Thief: 'Thief',
  Mage: 'Mage',
});

export const Color = Object.freeze({
  Black: 0, White: 1, Blue: 2, Red: 3,
  Purple: 4, Yellow: 5, Orange: 6, Green: 7,
});

export const COLOR_NAMES = ['Black', 'White', 'Blue', 'Red', 'Purple', 'Yellow', 'Orange', 'Green'];

export const CombatType = Object.freeze({
  Attack: 0, Block: 1, Parry: 2, Dodge: 3, Special: 4,
});

export const HitType = Object.freeze({
  None: 7,          // matches Go HitNone constant
  TheUser: 0,
  UserOnly: 1,
  UserMiss: 2,
  Opponent: 3,
  OpponentOnly: 4,
  Both: 5,
  Neither: 6,
  UserRan: 8,
  OpponentRan: 9,
});

export const WhereType = Object.freeze({
  TheTown: 0,
  TheInn: 1,
  TheWildernessMenu: 2,
  TheWildernessCombat: 3,
  TheHerbalist: 4,
  TheBank: 5,
  TheTavern: 6,
  TheTower: 7,
  TheJail: 8,
  TheBlacksmiths: 9,
  TheArenaMenu: 10,
  TheArenaCombat: 11,
  TheHealersHut: 12,
  TheCommonGuild: 13,
  TheFightersGuild: 14,
  TheThievesGuild: 15,
  TheMagesGuild: 16,
  TheArmory: 17,
});
