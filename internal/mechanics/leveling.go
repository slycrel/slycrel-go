package mechanics

import (
	"github.com/slycrel/slycrel/internal/model"
)

// NextLevelUp calculates the total experience required for a given level.
// Original: NextLevelUp() in Sly_CommonGuildUnit.cp
// Formula: f(x) = LevelUpDifficulty * x^3 + f(x-1), base = 50
func NextLevelUp(level int, difficulty int) int64 {
	if level <= 1 {
		return model.BaseExp
	}
	lvl := int64(level)
	return int64(difficulty)*lvl*lvl*lvl + NextLevelUp(level-1, difficulty)
}

// GiveNewLevel advances the character one level in their primary class.
// Original: GiveNewLevel() in Sly_CommonGuildUnit.cp
func GiveNewLevel(char *model.Character) {
	switch char.CharClass {
	case model.ClassFighter:
		char.FighterLvl++
		char.MaxHP += RandBetween(4, 5+char.FighterLvl)
		char.HitPoints = char.MaxHP
		char.Speed++
		char.Strength += RandBetween(1, 2)
		char.Dexterity += RandBetween(0, 1)
		char.Fame++
		char.Flirt1 += RandBetween(0, 2)
		char.Flirt2 += RandBetween(0, 2)

	case model.ClassMage:
		char.MageLvl++
		char.MaxHP += RandBetween(2, 3+char.MageLvl)
		char.HitPoints = char.MaxHP
		char.Speed++
		char.Strength += RandBetween(0, 1)
		char.Dexterity += RandBetween(0, 1)
		char.MaxPsyche += RandBetween(1, 2)
		char.Psyche = char.MaxPsyche
		char.Fame++
		char.Flirt1 += RandBetween(0, 2)
		char.Flirt2 += RandBetween(0, 2)

	case model.ClassThief:
		char.ThiefLvl++
		char.MaxHP += RandBetween(3, 4+char.ThiefLvl)
		char.HitPoints = char.MaxHP
		char.Speed += RandBetween(1, 2)
		char.Strength += RandBetween(1, 2)
		char.Dexterity += RandBetween(1, 2)
		char.Fame++
		char.Flirt1 += RandBetween(0, 2)
		char.Flirt2 += RandBetween(0, 2)
	}

	// Reset movement
	char.Movement = RandBetween(
		char.Speed-char.Speed/5,
		char.Speed+char.Speed/5,
	)
}
