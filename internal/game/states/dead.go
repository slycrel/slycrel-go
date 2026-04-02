package states

import (
	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
)

// DeadState handles player death - offers to start a new day or quit.
type DeadState struct{}

func (DeadState) ID() game.StateID { return "dead" }

func (DeadState) Enter(s *game.Session) {
	s.IO.Cr()
	choice := s.IO.LettersPrompt("[N]ew Day (resurrect & play again) or [Q]uit?", "NQ", 1, true, true)

	if choice == "N" {
		// Resurrect and start fresh day
		mechanics.NewDayForUser(s.Character)
		s.Character.LastOn = todayDate()
		s.Store.SaveCharacter(s.Character)
		s.IO.Cr()
		s.IO.Outln("A new day dawns... you have been restored.", true, 3)
		s.IO.Cr()
		s.SetNextState("begin")
	} else {
		s.IO.Outln("Now Returning to the BBS", true, 5)
		s.IO.Cr()
		s.Quit()
	}
}
