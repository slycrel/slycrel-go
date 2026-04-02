package states

import (
	"fmt"
	"math"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// healCost calculates the cost to heal a given amount of HP.
// Original formula from ST_AgathaHeals: cost = hpToHeal * floor(avgLevel * 1.25)
// Minimum cost of 1.
func healCost(char *model.Character, hpToHeal int) int {
	avg := (char.ThiefLvl + char.MageLvl + char.FighterLvl) / 3
	if avg < 2 {
		avg = 1
	}
	cost := hpToHeal * int(math.Trunc(float64(avg)*1.25))
	if cost < 1 && hpToHeal > 0 {
		cost = 1
	}
	return cost
}

// HealerState shows the healer ANSI menu.
// Original: ST_Healer
type HealerState struct{}

func (HealerState) ID() game.StateID { return "healer" }

func (HealerState) Enter(s *game.Session) {
	s.Character.Location = model.TheHealersHut
	s.IO.ShowANSIFile("healer_menu")
	s.SetNextState("healer_prompt")
}

// HealerPromptState shows HP/coins and prompts.
// Original: ST_HealerPrompt + ST_HealerMenu
type HealerPromptState struct{}

func (HealerPromptState) ID() game.StateID { return "healer_prompt" }

func (HealerPromptState) Enter(s *game.Session) {
	c := s.Character
	s.IO.Outln(fmt.Sprintf("You have %d/%d Hit Points and %d Coins on hand.",
		c.HitPoints, c.MaxHP, c.CoinsHand), true, 4)
	s.IO.Outln("[H]eal All, [N] Heal Amount, [Q]uit, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "HNQ?", 1, true, true)

	switch choice {
	case "H":
		s.SetNextState("agatha_heals")
	case "N":
		s.SetNextState("heal_amount")
	case "Q":
		s.Character.Location = model.TheWildernessMenu
		s.SetNextState("wilderness")
	case "?":
		s.SetNextState("healer")
	default:
		s.SetNextState("healer_prompt")
	}
}

// AgathaHealsState heals the player fully.
// Original: ST_AgathaHeals
type AgathaHealsState struct{}

func (AgathaHealsState) ID() game.StateID { return "agatha_heals" }

func (AgathaHealsState) Enter(s *game.Session) {
	c := s.Character
	hpNeeded := c.MaxHP - c.HitPoints

	if hpNeeded <= 0 {
		s.IO.Outln("You are already at full health!", true, 3)
		s.IO.Cr()
		s.SetNextState("healer_prompt")
		return
	}

	s.IO.Outln("Agatha waves her hands strangely above you...", true, 5)
	s.IO.Outln("..........", true, 1)
	s.IO.Cr()

	cost := healCost(c, hpNeeded)

	if c.CoinsHand < int64(cost) {
		s.IO.Cr()
		s.IO.Outln("Her eyes flutter open annoyed at you.", true, 5)
		s.IO.Outln("You don't have enough money, THIEF!!!", true, 6)
		s.IO.Cr()
	} else {
		c.CoinsHand -= int64(cost)
		c.HitPoints = c.MaxHP
		s.IO.Cr()
		s.IO.Outln(fmt.Sprintf("You have been HEALED! (Cost: %d coins)", cost), true, 3)
		s.IO.Cr()
	}

	s.Store.SaveCharacter(c)
	s.SetNextState("healer_prompt")
}

// HealAmountState heals a specific amount of HP.
// Original: ST_LifeHealedPrompt + ST_HealbyHealer
type HealAmountState struct{}

func (HealAmountState) ID() game.StateID { return "heal_amount" }

func (HealAmountState) Enter(s *game.Session) {
	c := s.Character
	hpNeeded := c.MaxHP - c.HitPoints

	if hpNeeded <= 0 {
		s.IO.Outln("You are already at full health!", true, 3)
		s.IO.Cr()
		s.SetNextState("healer_prompt")
		return
	}

	s.IO.Outln("How much life would you like healed?", true, 5)
	amount := s.IO.NumbersPrompt(":", 0, hpNeeded)

	if amount <= 0 {
		s.SetNextState("healer_prompt")
		return
	}

	cost := healCost(c, amount)

	if c.CoinsHand < int64(cost) {
		s.IO.Outln("What ARE you trying to do, rip me off???", true, 6)
		s.IO.Cr()
		s.SetNextState("healer_prompt")
	} else {
		c.CoinsHand -= int64(cost)
		c.HitPoints += amount
		if c.HitPoints > c.MaxHP {
			c.HitPoints = c.MaxHP
		}
		s.IO.Cr()
		s.IO.Outln(fmt.Sprintf("You have been HEALED! (+%d HP, Cost: %d coins)", amount, cost), true, 3)
		s.IO.Cr()
		s.Store.SaveCharacter(c)
	}

	s.SetNextState("healer_prompt")
}

// HerbalistState shows the herbalist menu.
// Original: ST_Herbalist
type HerbalistState struct{}

func (HerbalistState) ID() game.StateID { return "herbalist" }

func (HerbalistState) Enter(s *game.Session) {
	s.Character.Location = model.TheHerbalist
	s.IO.ShowANSIFile("herbalist_menu")
	s.SetNextState("herbalist_prompt")
}

// HerbalistPromptState shows stats and prompts.
// Original: ST_HerbalistPrompt + ST_HerbalistMenu
type HerbalistPromptState struct{}

func (HerbalistPromptState) ID() game.StateID { return "herbalist_prompt" }

func (HerbalistPromptState) Enter(s *game.Session) {
	c := s.Character
	s.IO.Outln(fmt.Sprintf("You have %d/%d Hit Points and %d Coins on hand.",
		c.HitPoints, c.MaxHP, c.CoinsHand), true, 4)
	s.IO.Outln("[H]eal, [Q]uit, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "HQ?", 1, true, true)

	switch choice {
	case "H":
		s.SetNextState("herbalist_heal")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "?":
		s.SetNextState("herbalist")
	default:
		s.SetNextState("herbalist_prompt")
	}
}

// HerbalistHealState heals the player (same mechanic as healer).
// Original: ST_HerbalistHeal
type HerbalistHealState struct{}

func (HerbalistHealState) ID() game.StateID { return "herbalist_heal" }

func (HerbalistHealState) Enter(s *game.Session) {
	c := s.Character
	hpNeeded := c.MaxHP - c.HitPoints

	if hpNeeded <= 0 {
		s.IO.Outln("You are already at full health!", true, 3)
		s.IO.Cr()
		s.SetNextState("herbalist_prompt")
		return
	}

	s.IO.Outln("The Herbalist waves her hands strangely above you...", true, 5)
	s.IO.Outln("..........", true, 1)
	s.IO.Cr()

	cost := healCost(c, hpNeeded)

	if c.CoinsHand < int64(cost) {
		s.IO.Cr()
		s.IO.Outln("Her eyes flutter open annoyed at you.", true, 2)
		s.IO.Outln("You don't have enough money, THIEF!!!", true, 6)
		s.IO.Cr()
	} else {
		c.CoinsHand -= int64(cost)
		c.HitPoints = c.MaxHP
		s.IO.Cr()
		s.IO.Outln(fmt.Sprintf("You have been HEALED!   <da da da da!> (Cost: %d coins)", cost), true, 3)
		s.IO.Cr()
	}

	s.Store.SaveCharacter(c)
	s.SetNextState("herbalist_prompt")
}
