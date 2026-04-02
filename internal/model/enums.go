package model

// WhereType represents all possible player locations.
type WhereType int

const (
	TheTown WhereType = iota
	TheInn
	TheWildernessMenu
	TheWildernessCombat
	TheHerbalist
	TheBank
	TheTavern
	TheTower
	TheJail
	TheBlacksmiths
	TheArenaMenu
	TheArenaCombat
	TheHealersHut
	TheCommonGuild
	TheFightersGuild
	TheThievesGuild
	TheMagesGuild
	TheArmory
)

func (w WhereType) String() string {
	names := [...]string{
		"Town", "Inn", "Wilderness", "Wilderness Combat",
		"Herbalist", "Bank", "Tavern", "Tower", "Jail",
		"Blacksmith", "Arena", "Arena Combat", "Healer's Hut",
		"Common Guild", "Fighter's Guild", "Thief's Guild",
		"Mage's Guild", "Armory",
	}
	if int(w) < len(names) {
		return names[w]
	}
	return "Unknown"
}

// TownType represents town sizes.
type TownType int

const (
	Village TownType = iota
	SmallTown
	Commonwealth
	Province
	City
	LargeCity
	Capitol
)

// HitType represents combat hit outcomes.
type HitType int

const (
	HitTheUser HitType = iota
	HitUserOnly
	HitUserMiss
	HitOpponent
	HitOpponentOnly
	HitBoth
	HitNeither
	HitNone
	HitUserRan
	HitOpponentRan
)

// CombatType represents combat action choices.
type CombatType int

const (
	CombatAttack CombatType = iota
	CombatBlock
	CombatParry
	CombatDodge
	CombatSpecial
)

// MonsterType represents creature categories.
type MonsterType int

const (
	Humanoid MonsterType = iota
	Orc
	Goblin
	Fairy
	Insectoid
	Fantastic
	Elven
	FElemental
	WElemental
	EElemental
	AElemental
	Dwarven
	Evil
	Good
	Flying
	Undead
	Psychic
	OtherMonster
)

func (m MonsterType) String() string {
	names := [...]string{
		"Humanoid", "Orc", "Goblin", "Fairy", "Insectoid",
		"Fantastic", "Elven", "Fire Elemental", "Water Elemental",
		"Earth Elemental", "Air Elemental", "Dwarven", "Evil",
		"Good", "Flying", "Undead", "Psychic", "Other",
	}
	if int(m) < len(names) {
		return names[m]
	}
	return "Unknown"
}

// Color represents favorite colors.
type Color int

const (
	ColorBlack Color = iota
	ColorWhite
	ColorBlue
	ColorRed
	ColorPurple
	ColorYellow
	ColorOrange
	ColorGreen
)

func (c Color) String() string {
	names := [...]string{
		"Black", "White", "Blue", "Red",
		"Purple", "Yellow", "Orange", "Green",
	}
	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// CharClass represents character occupations.
type CharClass string

const (
	ClassFighter CharClass = "Fighter"
	ClassThief   CharClass = "Thief"
	ClassMage    CharClass = "Mage"
)

// SpellsType represents mage spells.
type SpellsType int

const (
	SpellHeal SpellsType = iota
	SpellFlame
	SpellSnap
)

// ThiefSkillsType represents thief skills.
type ThiefSkillsType int

const (
	SkillPickPocket ThiefSkillsType = iota
	SkillPickLock
	SkillBackStab
)

// FighterSkillsType represents fighter skills.
type FighterSkillsType int

const (
	SkillLunge FighterSkillsType = iota
	SkillBreakWeapon
	SkillDisarm
)

// TerrainCell represents a terrain type on the combat grid.
type TerrainCell int

const (
	TerrainEmpty     TerrainCell = iota // 0 - Nothing/space
	TerrainPlainGr                      // 1 - Flat plain (green)
	TerrainPlainBr                      // 2 - Flat plain (brown/road)
	TerrainWater                        // 3 - Water (instant death)
	TerrainBridge                       // 4 - Bridge
	TerrainBoulder                      // 5 - Boulder
	TerrainForest                       // 6 - Light forest
	TerrainDeepForest                   // 7 - Deep forest
	TerrainSwamp                        // 8 - Light swamp
	TerrainDeepSwamp                    // 9 - Deep swamp
)
