package states

import (
	"github.com/slycrel/slycrel/internal/game"
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/mechanics"
)

// BeginState is the entry point - displays the opening screen and prompts to enter.
// Original: ST_BeginExternal + ST_GetOpeningANSI + ST_CheckForRegistration + ST_MainMenu
type BeginState struct{}

func (BeginState) ID() game.StateID { return "begin" }

func (BeginState) Enter(s *game.Session) {
	s.IO.ClearScreen()
	s.IO.ShowANSIFile("opening")
	s.IO.Cr()
	s.IO.Outln("=--=--  Welcome to the Realm of Slycrel  ---=--=", true, 3)
	s.IO.Cr()
	s.IO.Outln("[E]nter the Realm", true, 4)
	s.IO.Outln("[V]iew Guild Lists", true, 4)
	s.IO.Outln("[C]reate/View Character", true, 4)
	s.IO.Outln("[L]eave", true, 4)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your Selection >", "EVCL", 1, true, true)

	switch choice {
	case "E":
		s.SetNextState("enter_slycrel")
	case "V":
		// TODO: ViewGuildLists
		s.IO.Outln("Guild lists not yet implemented.", true, 1)
		s.IO.PausePrompt("--press a key--")
		s.SetNextState("begin")
	case "C":
		s.SetNextState("enter_slycrel") // will route to character creation if new
	case "L":
		s.IO.Outln("Exiting Slycrel...", true, 3)
		s.Quit()
	default:
		s.SetNextState("begin")
	}
}

// EnterSlycrelState handles loading or creating a character and entering town.
// Original: ST_EnterSlycrel
type EnterSlycrelState struct{}

func (EnterSlycrelState) ID() game.StateID { return "enter_slycrel" }

func (EnterSlycrelState) Enter(s *game.Session) {
	s.IO.Outln("-=---  "+s.Username+" Entered the realm of Slycrel.", true, 1)

	// Try to load existing character
	char, err := s.Store.FindCharacterByBBSName(s.Username)
	if err != nil {
		// New player - need to create character
		s.IO.Cr()
		s.IO.Outln("No character found. Creating a new one...", true, 3)
		s.SetNextState("create_character")
		return
	}

	s.Character = char

	// Check if dead today
	today := todayDate()
	if char.LastOn < today {
		// New day - reset daily limits and resurrect if dead
		if !char.Alive {
			s.IO.Outln("A new day dawns... you have been restored.", true, 3)
		}
		mechanics.NewDayForUser(char)
	} else if !char.Alive {
		s.IO.Outln("Sorry, You have been killed today. Please play again Tomorrow...", true, 6)
		s.IO.Cr()
		s.Character = char
		s.SetNextState("dead")
		return
	}

	// Check for mail/news
	news, _ := s.Store.ReadNews(s.Username)
	if news != "" {
		s.IO.Cr()
		s.IO.Outln("=== Daily News ===", true, 4)
		s.IO.Outln(news, true, 1)
		s.Store.ClearNews(s.Username)
		s.IO.PausePrompt("--press a key--")
	}

	s.SetNextState("town")
}

// QuitState handles saving and exiting.
// Original: ST_ShouldWeQuit + ST_WillWeQuit + ST_SaveAndQuit + ST_Quiting
type QuitState struct{}

func (QuitState) ID() game.StateID { return "quit" }

func (QuitState) Enter(s *game.Session) {
	s.IO.Cr()
	if !s.IO.YesNoQuestion("Are you SURE you wanna Leave?") {
		s.SetNextState("town_prompt")
		return
	}

	if s.Character != nil {
		s.Character.LastOn = todayDate()
		s.Store.SaveCharacter(s.Character)
		s.IO.Outln("Character saved.", true, 3)
	}

	s.IO.Outln("Now Returning to the BBS", true, 5)
	s.IO.Cr()
	s.Quit()
}

// todayDate returns today's date as YYYYMMDD integer.
func todayDate() int64 {
	now := slyio.Now()
	return int64(now.Year()*10000 + int(now.Month())*100 + now.Day())
}
