package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// JailState - random encounter at the Jail.
// Original: ST_Jail - 3 random outcomes. One kills you for the day!
type JailState struct{}

func (JailState) ID() game.StateID { return "jail" }

func (JailState) Enter(s *game.Session) {
	s.Character.Location = model.TheJail
	weapName := s.Character.Weapons[0].Name
	if weapName == "" {
		weapName = "weapon"
	}

	switch mechanics.RandBetween(0, 2) {
	case 0:
		s.IO.Outln("As you enter the jail, in a cell you see your fellow", true, 3)
		s.IO.Outln("hero Night Hawk. Quickly realizing this is no place", true, 3)
		s.IO.Outln("for a hero such as yourself, you head right back out", true, 3)
		s.IO.Outln("the door.", true, 3)
		s.IO.Cr()
		s.IO.PausePrompt("-=Press A Key=-")
		s.SetNextState("town")
	case 1:
		s.IO.Outln("As you enter the jail, the jailor walks up and shakes", true, 3)
		s.IO.Outln("your hand, saying he's always glad to see a hero visit.", true, 3)
		s.IO.Outln("The next thing you know you find yourself locked in a", true, 3)
		s.IO.Outln("cell with the jailor laughing at you from outside.", true, 3)
		s.IO.Outln("Looks like you'll have to spend the night.", true, 3)
		s.IO.Cr()
		s.Character.Alive = false
		s.Character.Location = model.TheTown
		s.Store.SaveCharacter(s.Character)
		s.IO.PausePrompt("-=Press A Key=-")
		s.SetNextState("dead")
	case 2:
		s.IO.Outln("As you enter the jail, the jailor walks up and shakes", true, 3)
		s.IO.Outln("your hand, saying he's glad to see such a fine hero", true, 3)
		s.IO.Outln(fmt.Sprintf("visit. The jailor takes you over to his weapon rack."), true, 3)
		s.IO.Outln(fmt.Sprintf("You can't help but notice the jailors %s", weapName), true, 3)
		s.IO.Outln("is in much better shape than your own. The jailor can't", true, 3)
		s.IO.Outln("help but notice your envy, and offers to loan you the", true, 3)
		s.IO.Outln("weapon. Before he can change his mind, you grab the", true, 3)
		s.IO.Outln("weapon and head out the door.", true, 3)
		s.IO.Cr()
		s.Character.Weapons[0].Strike += mechanics.RandBetween(1, 3)
		s.IO.PausePrompt("-=Press A Key=-")
		s.SetNextState("town")
	}
}
