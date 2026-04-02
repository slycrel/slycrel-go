package mechanics

import (
	"math"

	"github.com/slycrel/slycrel/internal/model"
)

// User2Monster converts a character into a monster record for arena combat.
// Original: User2Monster() in Sly_ArenaUnit.cp
func User2Monster(char *model.Character) model.Monster {
	weapName := char.Weapons[0].Name
	if weapName == "" {
		weapName = "fists"
	}
	weapon2Name := char.Weapons[1].Name
	if weapon2Name == "" {
		weapon2Name = "fists"
	}

	return model.Monster{
		Name:             char.Name,
		Level:            (char.ThiefLvl + char.MageLvl + char.FighterLvl) / 3,
		MonType:          model.Humanoid,
		MonsterNum:       RandBetween(1, 300),
		HitPoints:        char.HitPoints,
		Offense:          char.Strength/2 + char.Weapons[0].Strike,
		RangeOffense:     char.Strength/2 + char.Weapons[1].Strike,
		Range:            char.Weapons[1].Range,
		Defense:          (char.Dexterity+char.Speed)/3 + (char.Armor[0].Defense+char.Armor[1].Defense)/2,
		Movement:         RandBetween(char.Speed-char.Speed/5, char.Speed+char.Speed/5),
		Disposition:      1,
		Experience:       int(char.TotalExperience / 10),
		Coins:            int(math.Round(float64(char.CoinsHand) / 1.5)),
		AttackStr1:       weapName,
		AttackStr2:       weapon2Name,
		DefenseStr:       weapon2Name,
		DefenseActionStr: "blocks with",
	}
}

// ArenaAppraisal calculates a character's overall combat value for odds.
// Original: Arena_Appraisal / GetApprValue
func ArenaAppraisal(char *model.Character) int64 {
	val := char.TotalExperience +
		int64(char.ThiefLvl+char.MageLvl+char.FighterLvl)*100 +
		char.CoinsHand + char.CoinsBank

	fame := int64(char.Fame)
	if fame < 0 {
		fame = -fame
	}
	honor := int64(char.Honor)
	if honor < 0 {
		honor = -honor
	}
	faith := int64(char.Faith)
	if faith < 0 {
		faith = -faith
	}
	flirt := int64(char.Flirt1 + char.Flirt2)
	if flirt < 0 {
		flirt = -flirt
	}

	val += fame*100 + honor*100 + faith*100 + flirt*100
	return val
}

// CalcOdds returns the odds [challenger, opponent] for a gladiator fight.
// Original: ratio calculation in SetUpGCombat
func CalcOdds(challengerVal, opponentVal int64) [2]int {
	if challengerVal <= 0 || opponentVal <= 0 {
		return [2]int{1, 1}
	}

	ratio := float64(opponentVal) / float64(challengerVal)
	if ratio < 1 && ratio > 0 {
		invRatio := float64(challengerVal) / float64(opponentVal)
		return [2]int{1, int(math.Round(invRatio))}
	}
	return [2]int{int(math.Round(ratio)), 1}
}
