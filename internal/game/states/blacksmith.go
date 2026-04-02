package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// BlacksmithState - random encounter at Smith's shop.
// Original: ST_Blacksmith - 3 random outcomes affecting weapon Strike.
type BlacksmithState struct{}

func (BlacksmithState) ID() game.StateID { return "blacksmith" }

func (BlacksmithState) Enter(s *game.Session) {
	s.Character.Location = model.TheBlacksmiths
	weapName := s.Character.Weapons[0].Name
	if weapName == "" {
		weapName = "weapon"
	}

	switch mechanics.RandBetween(0, 2) {
	case 0:
		s.IO.Outln("You enter the blacksmith's shop and Smith immediately", true, 3)
		s.IO.Outln("begins laughing. Smith picks you up by the britches,", true, 3)
		s.IO.Outln("biceps bulging, and throws you back out on the road.", true, 3)
		s.IO.Cr()
	case 1:
		s.IO.Outln("You enter the blacksmith's shop and Smith gives you", true, 3)
		s.IO.Outln("a pat on the back. After you get back off the ground", true, 3)
		s.IO.Outln(fmt.Sprintf("and dust yourself off, Smith takes your %s", weapName), true, 3)
		s.IO.Outln("and begins to work on it. When he returns the weapon,", true, 3)
		s.IO.Outln("you can't help but notice its fine craftsmanship. You", true, 3)
		s.IO.Outln("immediately head for the door, thanking Smith as you leave.", true, 3)
		s.IO.Cr()
		s.Character.Weapons[0].Strike += mechanics.RandBetween(1, 3)
	case 2:
		s.IO.Outln("You enter the blacksmith's shop and Smith wallows up", true, 3)
		s.IO.Outln("to shake your hand. After you get back off the ground", true, 3)
		s.IO.Outln(fmt.Sprintf("and dust yourself off, Smith takes your %s", weapName), true, 3)
		s.IO.Outln("and begins to work on it. When he returns the weapon,", true, 3)
		s.IO.Outln("you can't help but notice it has several new dents.", true, 3)
		s.IO.Outln("You slowly head for the door. Smith smiles, and asks", true, 3)
		s.IO.Outln("you to come again soon.", true, 3)
		s.IO.Cr()
		s.Character.Weapons[0].Strike -= mechanics.RandBetween(1, 2)
		if s.Character.Weapons[0].Strike < 0 {
			s.Character.Weapons[0].Strike = 0
		}
	}

	s.IO.PausePrompt("-=Press A Key=-")
	s.SetNextState("town")
}
