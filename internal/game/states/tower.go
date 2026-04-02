package states

import (
	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// TowerState - random encounter in the Tower of Wisdom.
// Original: ST_Tower - 4 random outcomes affecting Faith.
type TowerState struct{}

func (TowerState) ID() game.StateID { return "tower" }

func (TowerState) Enter(s *game.Session) {
	s.Character.Location = model.TheTower

	switch mechanics.RandBetween(0, 3) {
	case 0:
		s.IO.Outln("You enter the dark dingy tower, and immediately encounter", true, 3)
		s.IO.Outln("a holy man saying a prayer in some strange language.", true, 3)
		s.IO.Outln("After he finishes, the holy man looks up and smiles,", true, 3)
		s.IO.Outln("happy to see a hero such as yourself. Again, he begins", true, 3)
		s.IO.Outln("to chant out a prayer....", true, 3)
		s.IO.Outln("Thankfully, its a blessing. When he's finished he", true, 3)
		s.IO.Outln("sends you on to do your hero's work.", true, 3)
		s.IO.Cr()
		s.Character.Faith += mechanics.RandBetween(1, 2)
	case 1:
		s.IO.Outln("You enter the tower, hopeful that you will gain some", true, 3)
		s.IO.Outln("much needed wisdom....", true, 3)
		s.IO.Cr()
		s.IO.Outln("Eventually, you encounter a High Priestess. The", true, 3)
		s.IO.Outln("priestess lets out a shriek of laugh, and casts a spell", true, 3)
		s.IO.Outln("on you...", true, 3)
		s.IO.Cr()
		s.IO.Outln("When you awaken, you feel dreadfully weak. When you", true, 3)
		s.IO.Outln("eventually find the strength, you leave the tower as", true, 3)
		s.IO.Outln("quickly as possible.", true, 3)
		s.IO.Cr()
		s.Character.Faith -= mechanics.RandBetween(1, 2)
	case 2:
		s.IO.Outln("As you enter the tower, you notice an evil feeling", true, 3)
		s.IO.Outln("has covered the area. Being a hero, you don't seem to mind.", true, 3)
		s.IO.Outln("After a considerable search, you encounter a wizard. The", true, 3)
		s.IO.Outln("Wizard immediately begins to cast a spell.....", true, 3)
		s.IO.Outln("Suddenly you find yourself back in town,", true, 3)
		s.IO.Outln("with an incredible amount of new energy.", true, 3)
		s.IO.Cr()
		s.Character.Faith += mechanics.RandBetween(1, 3)
	case 3:
		s.IO.Outln("You enter the tower, and a fearful feeling", true, 3)
		s.IO.Outln("immediately consumes your every thought. As you", true, 3)
		s.IO.Outln("peer through the darkness, you see a faint image", true, 3)
		s.IO.Outln("across the room. As you creep closer, you finally", true, 3)
		s.IO.Outln("realize you've encountered a High Priest. Before", true, 3)
		s.IO.Outln("you can even turn around, he casts a spell on you.", true, 3)
		s.IO.Outln("An incredible dizziness immediately overtakes you....", true, 3)
		s.IO.Cr()
		s.IO.Outln("You awaken in the center of town, so weak you can", true, 3)
		s.IO.Outln("barely stand.", true, 3)
		s.IO.Cr()
		s.Character.Faith -= mechanics.RandBetween(1, 2)
	}

	s.IO.PausePrompt("-=Press A Key=-")
	s.SetNextState("town")
}
