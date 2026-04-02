package states

import (
	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// TownState shows the town ANSI menu.
// Original: ST_Town
type TownState struct{}

func (TownState) ID() game.StateID { return "town" }

func (TownState) Enter(s *game.Session) {
	if s.Character != nil {
		s.Character.Location = model.TheTown
	}
	s.IO.ShowANSIFile("main_menu")
	s.SetNextState("town_prompt")
}

// TownPromptState shows the town menu prompt and routes to locations.
// Original: ST_TownShortPrompt + ST_GetTownMenu
type TownPromptState struct{}

func (TownPromptState) ID() game.StateID { return "town_prompt" }

func (TownPromptState) Enter(s *game.Session) {
	s.IO.Outln("[A,B,C,H,I,J,L,S,T,V,W,X,@,?]", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your Choice?", "ABCHIJLSTVWX@?", 1, true, true)

	switch choice {
	case "A":
		s.Character.Location = model.TheArenaMenu
		s.SetNextState("arena")
	case "B":
		s.Character.Location = model.TheBank
		s.SetNextState("bank")
	case "C":
		s.Character.Location = model.TheCommonGuild
		s.SetNextState("common_guild")
	case "H":
		s.Character.Location = model.TheHerbalist
		s.SetNextState("herbalist")
	case "I":
		inn, _ := s.Store.LoadInn()
		if inn != nil && !inn.Open && inn.Owner != s.Character.Name {
			s.IO.Outln("The Inn's windows are boarded up, with a sign painted on them--KEEP OUT!", true, 3)
			s.IO.Cr()
			s.IO.Outln("(This means that it's closed...)", true, 1)
			s.SetNextState("town_prompt")
		} else {
			s.Character.Location = model.TheInn
			s.SetNextState("inn")
		}
	case "J":
		s.Character.Location = model.TheJail
		s.SetNextState("jail")
	case "L":
		s.Character.Location = model.TheArmory
		s.SetNextState("armory")
	case "S":
		s.Character.Location = model.TheBlacksmiths
		s.SetNextState("blacksmith")
	case "T":
		s.Character.Location = model.TheTavern
		s.SetNextState("tavern")
	case "V":
		s.SetReturnState("town_prompt")
		s.SetNextState("view_character")
	case "W":
		s.Character.Location = model.TheWildernessMenu
		s.SetNextState("wilderness")
	case "X":
		s.SetNextState("quit")
	case "@":
		s.Character.Location = model.TheTower
		s.SetNextState("tower")
	case "?":
		s.SetNextState("town")
	default:
		s.SetNextState("town_prompt")
	}
}
