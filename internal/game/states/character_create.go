package states

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// CreateCharacterState handles initial character creation.
// Original: ST_CreateCharacter
type CreateCharacterState struct{}

func (CreateCharacterState) ID() game.StateID { return "create_character" }

func (CreateCharacterState) Enter(s *game.Session) {
	// Initialize a blank character
	char := &model.Character{
		BBSName:  s.Username,
		Name:     s.Username,
		Alive:    true,
		Gender:   true, // male default
		Location: model.TheTown,
	}
	s.Character = char

	// Random starting class
	switch rand.Intn(3) {
	case 0:
		rollChar(char, model.ClassFighter)
	case 1:
		rollChar(char, model.ClassThief)
	case 2:
		rollChar(char, model.ClassMage)
	}

	s.SetReturnState("create_character_menu")
	s.SetNextState("view_character")
}

// CreateCharMenuState shows the character creation menu.
// Original: ST_CreateCharacter menu display + ST_CreateCharMenu
type CreateCharMenuState struct{}

func (CreateCharMenuState) ID() game.StateID { return "create_character_menu" }

func (CreateCharMenuState) Enter(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln(" V)iew Character.", true, 4)
	s.IO.Outln(" N)ew Name.", true, 4)
	s.IO.Outln(" F)avorite Color", true, 4)
	s.IO.Outln(" C)hange Occupation.", true, 4)
	s.IO.Outln(" G)ender Change.", true, 4)
	s.IO.Outln(" R)eroll Attributes. (Caution! Resets to first level)", true, 4)
	s.IO.Outln(" S)how Scale.", true, 4)
	s.IO.Outln(" E)nter Town!", true, 4)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("[V,N,F,C,R,S,E,G]", "VNFCRSEG", 1, true, true)

	switch choice {
	case "V":
		s.SetReturnState("create_character_menu")
		s.SetNextState("view_character")
	case "N":
		s.SetNextState("get_char_name")
	case "F":
		s.SetNextState("favorite_color")
	case "C":
		s.SetNextState("occupation_menu")
	case "G":
		toggleGender(s)
		s.SetReturnState("create_character_menu")
		s.SetNextState("view_character")
	case "R":
		rollChar(s.Character, s.Character.CharClass)
		s.SetReturnState("create_character_menu")
		s.SetNextState("view_character")
	case "S":
		showScale(s)
		s.SetNextState("create_character_menu")
	case "E":
		s.SetNextState("enter_town_from_create")
	default:
		s.SetNextState("create_character_menu")
	}
}

// EnterTownFromCreateState saves the new character and enters the game.
type EnterTownFromCreateState struct{}

func (EnterTownFromCreateState) ID() game.StateID { return "enter_town_from_create" }

func (EnterTownFromCreateState) Enter(s *game.Session) {
	// Check if name is taken by someone else
	existing, err := s.Store.FindCharacterByName(s.Character.Name)
	if err == nil && !strings.EqualFold(existing.BBSName, s.Character.BBSName) {
		s.IO.Outln("Sorry... that name has already been taken.", true, 1)
		s.IO.Outln("Please choose another name.", true, 1)
		s.SetNextState("get_char_name")
		return
	}

	// Check if this is a new or existing player
	_, err = s.Store.FindCharacterByBBSName(s.Character.BBSName)
	if err != nil {
		// New player
		if addErr := s.Store.AddCharacter(s.Character); addErr != nil {
			s.IO.Outln("Error saving character: "+addErr.Error(), true, 6)
		}
	} else {
		// Existing player edited their character
		s.Store.SaveCharacter(s.Character)
	}

	s.SetNextState("enter_slycrel")
}

// GetCharNameState prompts for a new character name.
type GetCharNameState struct{}

func (GetCharNameState) ID() game.StateID { return "get_char_name" }

func (GetCharNameState) Enter(s *game.Session) {
	name := s.IO.ReadLine("What is your new Name:", 20)
	if name != "" {
		s.Character.Name = name
	}
	s.SetNextState("create_character_menu")
}

// OccupationMenuState prompts for character class choice.
type OccupationMenuState struct{}

func (OccupationMenuState) ID() game.StateID { return "occupation_menu" }

func (OccupationMenuState) Enter(s *game.Session) {
	s.IO.Outln("[F]ighter", true, 1)
	s.IO.Outln("[T]hief", true, 1)
	s.IO.Outln("[M]age", true, 1)

	choice := s.IO.LettersPrompt("New class -", "FTM", 1, true, true)

	switch choice {
	case "F":
		rollChar(s.Character, model.ClassFighter)
	case "T":
		rollChar(s.Character, model.ClassThief)
	case "M":
		rollChar(s.Character, model.ClassMage)
	}

	s.SetReturnState("create_character_menu")
	s.SetNextState("view_character")
}

// FavoriteColorState prompts for favorite color.
type FavoriteColorState struct{}

func (FavoriteColorState) ID() game.StateID { return "favorite_color" }

func (FavoriteColorState) Enter(s *game.Session) {
	s.IO.Cr()
	// Palette is the 6-color IOProvider palette; black/orange map to the
	// nearest available hue since the protocol doesn't carry true RGB.
	s.IO.Outln(" 1 - Black", true, 2)   // no black in palette → white-ish
	s.IO.Outln(" 2 - White", true, 2)
	s.IO.Outln(" 3 - Blue", true, 1)    // cyan is closest to blue
	s.IO.Outln(" 4 - Red", true, 6)
	s.IO.Outln(" 5 - Purple", true, 5)
	s.IO.Outln(" 6 - Yellow", true, 4)
	s.IO.Outln(" 7 - Orange", true, 4)  // yellow-ish stand-in
	s.IO.Outln(" 8 - Green", true, 3)
	s.IO.Cr()

	n := s.IO.NumbersPrompt("# of Color :", 1, 8)
	s.Character.FavColor = model.Color(n - 1)
	s.SetNextState("create_character_menu")
}

// ViewCharacterState displays the character sheet and returns.
// Original: ST_ViewCharacter -> PrintCharStats
type ViewCharacterState struct{}

func (ViewCharacterState) ID() game.StateID { return "view_character" }

func (ViewCharacterState) Enter(s *game.Session) {
	c := s.Character
	if c == nil {
		s.IO.Outln("No character to display.", true, 6)
		s.PopReturnState()
		return
	}

	s.IO.Cr()
	s.IO.Outln("=--=-- Character Sheet ---=--=", true, 3)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Name:       %s", c.Name), true, 4)
	s.IO.Outln(fmt.Sprintf("  Class:      %s", c.CharClass), true, 1)
	s.IO.Outln(fmt.Sprintf("  Gender:     %s", c.GenderString()), true, 1)
	s.IO.Outln(fmt.Sprintf("  Level:      F:%d  T:%d  M:%d", c.FighterLvl, c.ThiefLvl, c.MageLvl), true, 1)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Hit Points: %d/%d", c.HitPoints, c.MaxHP), true, 2)
	s.IO.Outln(fmt.Sprintf("  Psyche:     %d/%d", c.Psyche, c.MaxPsyche), true, 2)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Strength:   %d", c.Strength), true, 1)
	s.IO.Outln(fmt.Sprintf("  Dexterity:  %d", c.Dexterity), true, 1)
	s.IO.Outln(fmt.Sprintf("  Speed:      %d", c.Speed), true, 1)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Coins:      %d (Bank: %d)", c.CoinsHand, c.CoinsBank), true, 4)
	s.IO.Outln(fmt.Sprintf("  Experience: %d (Spending: %d)", c.TotalExperience, c.SpendingExperience), true, 1)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Fame: %d  Honor: %d  Faith: %d", c.Fame, c.Honor, c.Faith), true, 1)
	s.IO.Outln(fmt.Sprintf("  Color:      %s", c.FavColor), true, 1)

	// Weapons
	for i, w := range c.Weapons {
		if !w.IsEmpty() {
			s.IO.Outln(fmt.Sprintf("  Weapon %d:   %s (Strike:%d Range:%d)", i+1, w.Name, w.Strike, w.Range), true, 1)
		}
	}
	// Armor
	for i, a := range c.Armor {
		if !a.IsEmpty() {
			s.IO.Outln(fmt.Sprintf("  Armor %d:    %s (Defense:%d)", i+1, a.Name, a.Defense), true, 1)
		}
	}

	s.IO.Cr()
	s.PopReturnState()
}

// rollChar initializes a character's stats for the given class.
// Faithfully recreates RollChar from Sly_MainUnit.cp.
func rollChar(c *model.Character, class model.CharClass) {
	c.CharClass = class
	c.FighterLvl = 1
	c.ThiefLvl = 1
	c.MageLvl = 1
	c.TotalExperience = model.BaseExp
	c.SpendingExperience = model.BaseExp
	c.CoinsHand = model.StartingCoinsHand
	c.CoinsBank = model.StartingCoinsBank
	c.Alive = true
	c.Fame = 0
	c.Honor = 0
	c.Faith = 0
	c.Flirt1 = 0
	c.Flirt2 = 0
	c.TotalFights = 0
	c.FightsWon = 0
	c.Exploration = model.MaxExploration
	c.Spars = model.MaxSpars

	switch class {
	case model.ClassFighter:
		// High HP (15-35), Strength 4-12, low magic
		c.MaxHP = randBetween(15, 35)
		c.Strength = randBetween(4, 12)
		c.Dexterity = randBetween(4, 12)
		c.Speed = randBetween(6, 12)
		c.MaxPsyche = 1
		c.Movement = randBetween(8, 14)
	case model.ClassThief:
		// Speed 10-16, Dexterity up to 20, low HP
		c.MaxHP = randBetween(13, 26)
		c.Strength = randBetween(3, 10)
		c.Dexterity = randBetween(8, 20)
		c.Speed = randBetween(10, 16)
		c.MaxPsyche = 1
		c.Movement = randBetween(10, 18)
	case model.ClassMage:
		// Lowest HP, highest magic, lowest strength
		c.MaxHP = randBetween(4, 17)
		c.Strength = randBetween(2, 10)
		c.Dexterity = randBetween(4, 14)
		c.Speed = randBetween(6, 12)
		c.MaxPsyche = randBetween(2, 4)
		c.Movement = randBetween(6, 12)
	}

	c.HitPoints = c.MaxHP
	c.Psyche = c.MaxPsyche

	// Starting weapon based on class
	switch class {
	case model.ClassFighter:
		c.Weapons[0] = model.Weapon{Name: "Short Sword", Strike: 3, Range: 0, ActionStr: "slash"}
	case model.ClassThief:
		c.Weapons[0] = model.Weapon{Name: "Dagger", Strike: 2, Range: 1, ActionStr: "stab"}
	case model.ClassMage:
		c.Weapons[0] = model.Weapon{Name: "Staff", Strike: 1, Range: 0, ActionStr: "strike"}
	}
}

// toggleGender switches gender and applies stat modifiers.
// Original: case 'G' in ST_CreateCharMenu
func toggleGender(s *game.Session) {
	c := s.Character
	if c.Gender {
		// Male -> Female
		s.IO.Outln("You are Now Female.", true, 3)
		c.Gender = false
		c.Strength -= 3
		c.MaxPsyche++
		c.Dexterity += 2
		c.MaxHP -= 2
		c.HitPoints = c.MaxHP
		c.Faith += 2
	} else {
		// Female -> Male
		s.IO.Outln("You are Now Male.", true, 3)
		c.Gender = true
		c.Strength += 3
		c.MaxPsyche--
		c.Dexterity -= 2
		c.MaxHP += 2
		c.HitPoints = c.MaxHP
		c.Faith -= 2
	}
}

// showScale displays the stat scale.
func showScale(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln("=--=-- Attribute Scale ---=--=", true, 3)
	s.IO.Outln("  1-3   : Poor", true, 6)
	s.IO.Outln("  4-6   : Below Average", true, 4)
	s.IO.Outln("  7-9   : Average", true, 1)
	s.IO.Outln("  10-12 : Above Average", true, 1)
	s.IO.Outln("  13-15 : Good", true, 3)
	s.IO.Outln("  16-18 : Excellent", true, 3)
	s.IO.Outln("  19+   : Exceptional", true, 3)
	s.IO.Cr()
	s.IO.PausePrompt("--press a key--")
}

// randBetween returns a random integer in [min, max] inclusive.
func randBetween(min, max int) int {
	if min >= max {
		return min
	}
	return min + rand.Intn(max-min+1)
}
