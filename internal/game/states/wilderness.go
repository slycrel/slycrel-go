package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// WildernessState shows the wilderness ANSI menu.
// Original: ST_Wilderness
type WildernessState struct{}

func (WildernessState) ID() game.StateID { return "wilderness" }

func (WildernessState) Enter(s *game.Session) {
	s.Character.Location = model.TheWildernessMenu
	s.IO.ShowANSIFile("wilderness_menu")
	s.SetNextState("wilderness_prompt")
}

// WildernessPromptState shows the wilderness menu and routes choices.
// Original: ST_WildernessShortPrompt + ST_WildernessMenu
type WildernessPromptState struct{}

func (WildernessPromptState) ID() game.StateID { return "wilderness_prompt" }

func (WildernessPromptState) Enter(s *game.Session) {
	s.IO.Outln(fmt.Sprintf("  Explorations remaining: %d", s.Character.Exploration), true, 1)
	s.IO.Outln("[F]orest, [M]ountain, [S]wamp, [A]gatha's Hut, [Q]uit to Town, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your selection?", "AFMSQ?", 1, true, true)

	switch choice {
	case "A":
		s.Character.Location = model.TheHealersHut
		s.SetNextState("healer")
	case "F", "M", "S":
		s.Character.Location = model.TheWildernessCombat
		// Store the region choice for combat setup
		switch choice {
		case "F":
			s.CombatRegion = "forest"
		case "M":
			s.CombatRegion = "mountain"
		case "S":
			s.CombatRegion = "swamp"
		}
		s.SetNextState("setup_combat")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "?":
		s.SetNextState("wilderness")
	default:
		s.SetNextState("wilderness_prompt")
	}
}

// SetupCombatState loads a monster and initiates combat.
// Original: ST_SetupCombat
type SetupCombatState struct{}

func (SetupCombatState) ID() game.StateID { return "setup_combat" }

func (SetupCombatState) Enter(s *game.Session) {
	s.Character.TotalFights++

	if s.Character.Exploration <= 0 {
		s.IO.Outln("You just can't find any more monsters today... Try again tomorrow.", true, 3)
		s.IO.Cr()
		s.IO.PausePrompt("-=Press A Key=-")
		s.SetNextState("wilderness")
		return
	}

	s.Character.Exploration--

	// Load a random monster from the chosen region
	monster, err := s.Store.GetRandomMonster(s.CombatRegion, s.Character.Level())
	if err != nil {
		s.IO.Outln("The wilderness is eerily quiet... no monsters found.", true, 3)
		s.IO.PausePrompt("--press a key--")
		s.SetNextState("wilderness")
		return
	}

	s.Monster = monster
	s.MonHitPoints = monster.HitPoints // save max HP for disposition calc

	// Reset movement for this encounter
	spd := s.Character.Speed
	if spd < 5 {
		spd = 5
	}
	s.Character.Movement = randBetween(spd-spd/5, spd+spd/5)

	s.IO.ClearScreen()
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("You encounter a %s!", monster.Name), true, 6)
	s.IO.Cr()

	// Try grid combat first (if terrain maps available), fall back to text combat
	s.SetNextState("grid_combat_setup")
}
