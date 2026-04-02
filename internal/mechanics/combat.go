package mechanics

import (
	"math"
	"math/rand"

	"github.com/slycrel/slycrel/internal/model"
)

// RollInitiative determines whether the player goes first.
// Original: RollInitiative() in Sly_TextCombatUnit.cp
// Returns true if the player has initiative.
func RollInitiative(playerMovement, monsterMovement int) bool {
	diff := playerMovement - monsterMovement
	return diff+rand.Intn(4) >= 0
}

// DoDamageUser calculates damage dealt by the player to a monster.
// Original: DoDamage(true) in Sly_TextCombatUnit.cp
func DoDamageUser(char *model.Character, monster *model.Monster) int {
	avgLevel := (char.ThiefLvl + char.MageLvl + char.FighterLvl) / 3

	var damage int
	if avgLevel <= 4 {
		// Low level formula
		low := int(math.Round(float64(char.Strength)*0.75+float64(char.Weapons[0].Strike)) * 1.25)
		high := int(math.Round((float64(char.Strength)/0.75 + float64(char.Weapons[0].Strike)) * 1.35))
		damage = RandBetween(low, high) - monster.Defense
		if damage < 2 {
			damage = rand.Intn(3) // 0-2
		}
	} else {
		// High level formula
		low := int(math.Round(float64(char.Strength)/2 + float64(char.Weapons[0].Strike)))
		high := int(math.Round((float64(char.Strength)/2 + float64(char.Weapons[0].Strike)) * 1.25))
		damage = RandBetween(low, high) - monster.Defense
	}

	if damage < 0 {
		damage = 0
	}
	return damage
}

// DoDamageMonster calculates damage dealt by a monster to the player.
// Original: DoDamage(false) in Sly_TextCombatUnit.cp
func DoDamageMonster(char *model.Character, monster *model.Monster) int {
	low := int(math.Trunc(float64(monster.Offense) * 0.35))
	high := int(math.Trunc(float64(monster.Offense) * 1.35))
	damage := RandBetween(low, high) - char.Armor[0].Defense

	if damage < 0 {
		damage = 0
	}
	return damage
}

// RandMonsterAction picks a random combat action for the monster.
// Original: rndMonsterAttack() in Sly_TextCombatUnit.cp
// Returns the action and whether the monster ran.
func RandMonsterAction(monsterHP, monsterMaxHP int) (model.CombatType, bool) {
	roll := RandBetween(1, 50)

	// Check if monster wants to run (low HP)
	wantsToRun := monsterMaxHP/20 > monsterHP

	switch {
	case roll <= 25:
		return model.CombatAttack, false
	case roll <= 30:
		return model.CombatParry, false
	case roll <= 37:
		return model.CombatBlock, false
	case roll <= 44:
		return model.CombatDodge, false
	default: // 45-50
		if wantsToRun {
			return model.CombatAttack, true // monster runs
		}
		return model.CombatAttack, false
	}
}

// CalcAdvanceMoves calculates how many extra attacks a combatant gets
// based on speed advantage in the combat loop.
// Original: advance moves calculation in ST_CalcAdvanceMoves
func CalcAdvanceMoves(attackerLoop, defenderLoop, combatLoop int) int {
	moves := 0
	loop := combatLoop
	for {
		loop--
		if loop <= 0 {
			break
		}
		if attackerLoop >= loop && defenderLoop < loop {
			moves++
		} else if (attackerLoop >= loop && defenderLoop >= loop) ||
			(attackerLoop < loop && defenderLoop >= loop) {
			break
		}
	}
	return moves
}

// ApplyUserAttackModifier adjusts damage based on the user's action vs monster's action.
// Original: modifier logic in CalcUserMove
func ApplyUserAttackModifier(baseDamage int, monMove model.CombatType) int {
	switch monMove {
	case model.CombatBlock:
		return baseDamage / 2
	case model.CombatParry:
		return int(math.Round(float64(baseDamage) * 0.6))
	default:
		return baseDamage
	}
}

// ApplyMonsterAttackModifier adjusts monster damage based on the player's action.
// Original: modifier logic in MonsterAttacks
func ApplyMonsterAttackModifier(baseDamage int, userMove model.CombatType, playerDex, monsterMove int) int {
	switch userMove {
	case model.CombatBlock:
		return baseDamage / 2
	case model.CombatParry:
		return int(math.Round(float64(baseDamage) * 0.6))
	case model.CombatSpecial:
		return int(math.Round(float64(baseDamage) * 1.75))
	case model.CombatDodge:
		if RandBetween(1, monsterMove) < playerDex {
			return 0
		}
		return baseDamage
	default:
		return baseDamage
	}
}

// Disposition represents the relative state of both combatants.
type Disposition int

const (
	DispMgoodUgood Disposition = iota + 1
	DispMgoodUbad
	DispMbadUgood
	DispMokUok
	DispMokUgood
	DispMokUbad
	DispMgoodUok
	DispMbadUok
	DispMbadUbad
)

// GetDisposition calculates the 9-state disposition based on HP percentages.
// Original: GetMainCombatText disposition calculation in Sly_TextCombatUnit.cp
func GetDisposition(userHP, userMax, monHP, monMax int) Disposition {
	var userPct, monPct int
	if userMax > 0 && userHP > 0 {
		userPct = int(math.Round(float64(userHP) / float64(userMax) * 100))
	}
	if monMax > 0 && monHP > 0 {
		monPct = int(math.Round(float64(monHP) / float64(monMax) * 100))
	}

	if userPct >= 70 {
		if monPct >= 70 {
			return DispMgoodUgood
		} else if monPct >= 35 {
			return DispMokUgood
		}
		return DispMbadUgood
	} else if userPct >= 35 {
		if monPct >= 70 {
			return DispMgoodUok
		} else if monPct >= 35 {
			return DispMokUok
		}
		return DispMbadUok
	} else {
		if monPct >= 70 {
			return DispMgoodUbad
		} else if monPct >= 35 {
			return DispMokUbad
		}
		return DispMbadUbad
	}
}

// RandBetween returns a random integer in [min, max] inclusive.
func RandBetween(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}
