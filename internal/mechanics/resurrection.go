package mechanics

import (
	"math"

	"github.com/slycrel/slycrel/internal/model"
)

// ResurrectUser applies death penalties and brings the character back to life.
// Original: ResurrectUser in General_Unit.cp
// Penalties: -2% TotalExperience, -2% SpendingExperience, -1 Fame, +1 Faith, -20% CoinsHand
func ResurrectUser(char *model.Character) {
	char.TotalExperience -= int64(math.Round(float64(char.TotalExperience) * 0.02))
	char.SpendingExperience -= int64(math.Round(float64(char.SpendingExperience) * 0.02))
	char.Fame--
	char.Faith++
	char.CoinsHand -= int64(math.Round(float64(char.CoinsHand) * 0.20))

	if char.TotalExperience < 0 {
		char.TotalExperience = 0
	}
	if char.SpendingExperience < 0 {
		char.SpendingExperience = 0
	}
	if char.CoinsHand < 0 {
		char.CoinsHand = 0
	}

	char.Alive = true
	char.HitPoints = char.MaxHP
}

// NewDayForUser resets daily limits and restores the character for a new day.
// Original: NewDayForUser in Sly_MainUnit.cp
func NewDayForUser(char *model.Character) {
	char.Psyche = char.MaxPsyche
	char.TotalExperience += 15
	char.SpendingExperience += 15
	char.Movement = RandBetween(
		char.Speed-char.Speed/5,
		char.Speed+char.Speed/5,
	)
	char.Exploration = model.MaxExploration
	char.Spars = model.MaxSpars

	if !char.Alive {
		ResurrectUser(char)
	} else {
		char.HitPoints = char.MaxHP
	}
}
