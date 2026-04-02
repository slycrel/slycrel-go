package states

import (
	"fmt"
	"math/rand"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// Drink costs indexed by drink number (1-15). Index 0 unused.
var drinkCosts = [16]int{0, 0, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 7, 9, 10, 10}

var drinkTexts = [16]string{
	"",
	"Lynx gives you a hard look and shifts his eyes slightly.\n\"Here. But don't let anyone know I still have this stuff...\"",
	"The bartender suppresses a smile as he hands you the drink.",
	"Lynx hands you an ale. When you don't tip him, he gets an\nangry look on his face, and stalks off.",
	"Lynx hands you the drink. You think he winks as you turn,\nbut you can't be certain.",
	"Lynx stirs the cocktail slightly and hands it to you.\n\nEWW!!! He licked the spoon and stuck it in that guy's drink!",
	"Mmmm! Delicious. (You hope no faeries were still in it...)",
	"The Bartender hands you the wine, and tells you:\n\"That woman over in the corner... I think she likes you!\"",
	"Lynx pours whiskey into a small glass. He licks the rim of\nthe bottle as he walks off.",
	"The bartender tells you this stuff's expensive, and how hard it's\ngoing to be to replace.\n\nYou just shrug, and gleefully chug down your drink.",
	"You shudder as the life essence of the poor, innocent,\nunsuspecting little dwarf rolls down your throat.",
	"The bartender steps into the back for a moment. When he\nreturns, he is carrying a slightly thicker mug with a lid.\n\"I ain't responsible,\" he tells you as he hands it over.\n\nThe drink oozes down your throat, nearly suffocating you.",
	"The barkeep reaches under the counter, and pulls out a large, black\njug. He pours you a tiny glass.\n\nThe drink burns as it goes down.",
	"The bartender hands you a steaming glass of amber liquid.\n\nThe drink runs smoothly down your throat.\nYou suddenly feel the need to use the restroom.",
	"The bartender laughs, and says, \"You HONESTLY thought that\nI'd even let Josepi IN here?\" He throws back his head and laughs.\n\nwell, you DID ask... (And pay, too!)",
	"Lynx shrugs as if to say \"hey, it's not my life\" and\nhands it to you.\n\nYou suddenly feel as if the world turns upside down.\nYou sway and pass out...\n\nYou'd better check your stats bud!!!",
}

var flirtFailTexts = [9]string{
	"",
	"Corenne doesn't seem to see you, but Grego perks right up\nand runs over, eagerly awaiting your order...",
	"You take her hand as she turns away, but just as you lift it to your\nlips, she jerks free, quickly walking to a much better looking customer.",
	"You bravely announce that Corenne has the best looking elbows that\nyou have ever seen. She goes red all over and you congratulate\nyourself. You do wonder, however, why Lynx and the other bar\npatrons are laughing at you...",
	"You stand, gathering everyone's attention with a manly <ahem>.\nSomeone in the back snickers as the bar goes silent. You start\nrambling off miscellaneous heroic deeds that you have done. After\nless than a minute, you feel slightly sick. You wonder at fate,\nand all its quirks as you spill your lunch on the floor...",
	"You very obviously wink seductively at Corenne several times.\nShe hurries over to you and asks if your eye is okay. You\nnearly choke as you mumble you're fine...",
	"You compliment Corenne's attributes and how bouncy they are.\nYou suddenly find yourself staring at the ceiling with a\nsplit lip.",
	"As Corenne walks by, you ask her to dance. She stares at you\nfor a moment, and reminds you that there is no music playing.\nGrego quickly snakes in, and informs you he will dance.\n\nYou shudder at the thought...",
	"You ask Corenne if she would like to sit a while and share a meal.\nShe replies that she is on a diet, and that she needs to keep\nher figure for the MEN...\n\nYour pride is hurt....",
}

// TavernState shows the tavern ANSI menu.
type TavernState struct{}

func (TavernState) ID() game.StateID { return "tavern" }

func (TavernState) Enter(s *game.Session) {
	s.Character.Location = model.TheTavern
	s.IO.ShowANSIFile("tavern_menu")
	s.SetNextState("tavern_prompt")
}

// TavernPromptState shows the tavern menu.
type TavernPromptState struct{}

func (TavernPromptState) ID() game.StateID { return "tavern_prompt" }

func (TavernPromptState) Enter(s *game.Session) {
	s.IO.Outln("[F]lirt, [H]ang Around, [L]isten, [O]rder Drink, [Q]uit, [T]alk to Lynx, [V]iew Guilds, [?]", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Well? :", "FHLOQTV?", 1, true, true)

	switch choice {
	case "F":
		s.SetNextState("corenne_prompt")
	case "H":
		s.SetNextState("hang_around")
	case "L":
		s.IO.Outln("You sit on a bar stool for a while trying to listen", true, 2)
		s.IO.Outln("to every whisper from across the room. After some time", true, 2)
		s.IO.Outln("you decide you can't hear anything and wonder if you", true, 2)
		s.IO.Outln("should just leave the tavern.", true, 2)
		s.IO.Cr()
		s.SetNextState("tavern_prompt")
	case "O":
		s.SetNextState("order_drink")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "T":
		s.SetNextState("talk_to_lynx")
	case "V":
		s.SetNextState("view_guilds")
	case "?":
		s.SetNextState("tavern")
	default:
		s.SetNextState("tavern_prompt")
	}
}

// OrderDrinkState shows the drink menu and processes the order.
type OrderDrinkState struct{}

func (OrderDrinkState) ID() game.StateID { return "order_drink" }

func (OrderDrinkState) Enter(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln("Whaddya want?", true, 1)
	s.IO.Cr()
	s.IO.Outln(" 1- Fruit Juice                  0", true, 2)
	s.IO.Outln(" 2- Elven Ale                    1", true, 2)
	s.IO.Outln(" 3- Ale                          2", true, 2)
	s.IO.Outln(" 4- Mead                         2", true, 2)
	s.IO.Outln(" 5- Cocktail                     3", true, 2)
	s.IO.Outln(" 6- Faery Water                  3", true, 2)
	s.IO.Outln(" 7- Fine Wine                    4", true, 2)
	s.IO.Outln(" 8- Whiskey                      4", true, 2)
	s.IO.Outln(" 9- Ogre Sweat                   5", true, 2)
	s.IO.Outln("10- Dwarf Spirits                5", true, 2)
	s.IO.Outln("11- Dragon Blood                 6", true, 2)
	s.IO.Outln("12- Potency                      7", true, 2)
	s.IO.Outln("13- Kwick                        9", true, 2)
	s.IO.Outln("14- Josepi's Molotov Cocktail   10", true, 2)
	s.IO.Outln("15- Auneletho's Surprise        10", true, 2)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("$%d in Hand", s.Character.CoinsHand), true, 3)
	s.IO.Cr()

	num := s.IO.NumbersPrompt("[1-15, 0=Quit]:", 0, 16)

	if num == 0 {
		s.SetNextState("tavern_prompt")
		return
	}

	// Easter egg: ordering 16 kills you (from original)
	if num == 16 {
		s.IO.Outln("You just can't follow directions, can you? I said 1 to 15, NOT 16...", true, 6)
		s.IO.Outln("But that's okay. I won't mind killing you.", true, 2)
		s.IO.Cr()
		s.IO.Outln("                   <ZZZZZZZZZZZZZZZT!!!>", true, 4)
		s.Character.Alive = false
		s.Store.SaveCharacter(s.Character)
		s.SetNextState("dead")
		return
	}

	cost := drinkCosts[num]
	if s.Character.CoinsHand < int64(cost) {
		s.IO.Outln("You can't afford that drink!", true, 6)
		s.IO.Cr()
	} else {
		s.Character.CoinsHand -= int64(cost)
		s.IO.Cr()
		s.IO.Outln(drinkTexts[num], true, rand.Intn(8))
		s.IO.Cr()

		// Fruit juice special effect: reduce flirt cooldown
		if num == 1 {
			s.Character.Flirt1--
			if s.Character.Flirt1 < 1 {
				s.Character.Flirt1 = 1
			}
		}
	}

	s.IO.Cr()
	s.SetNextState("tavern_prompt")
}

// TalkToLynxState - Lynx is busy.
type TalkToLynxState struct{}

func (TalkToLynxState) ID() game.StateID { return "talk_to_lynx" }

func (TalkToLynxState) Enter(s *game.Session) {
	s.IO.Outln("Sorry, Lynx doesn't feel up to talking right now. He's busy", true, rand.Intn(8))
	s.IO.Outln("playing with the barmaid...", true, rand.Intn(8))
	s.IO.Cr()
	s.SetNextState("tavern_prompt")
}

// HangAroundState - lounge around the bar.
type HangAroundState struct{}

func (HangAroundState) ID() game.StateID { return "hang_around" }

func (HangAroundState) Enter(s *game.Session) {
	s.IO.Outln("You lounge around the bar for a while.", true, rand.Intn(8))
	s.IO.Cr()

	if mechanics.RandBetween(1, 100) <= 50 {
		s.IO.Outln("This is boring. Let's do something else.", true, rand.Intn(8))
		s.IO.Cr()
		s.IO.PausePrompt("ZZZZZzzzzzz...")
	} else {
		s.IO.Outln("You nod off after a while....", true, rand.Intn(8))
		s.IO.Cr()
		s.IO.Outln("ZZzzzZZzzzZZ", true, 1)
		s.IO.Cr()
		s.IO.Outln("As you wake up, your money purse feels lighter... You must have spent more than you thought...", true, 2)
		divisor := mechanics.RandBetween(10, 25)
		if divisor > 0 {
			s.Character.CoinsHand -= s.Character.CoinsHand / int64(divisor)
		}
	}

	s.SetNextState("tavern_prompt")
}

// CorennePromptState - flirt menu.
type CorennePromptState struct{}

func (CorennePromptState) ID() game.StateID { return "corenne_prompt" }

func (CorennePromptState) Enter(s *game.Session) {
	if s.Character.Flirt1 < 0 {
		s.IO.Outln("Corenne will probably think you're desperate, and you wouldn't want that.", true, 3)
		s.SetNextState("tavern_prompt")
		return
	}

	s.IO.Cr()
	s.IO.Outln("1] Grin Foolishly", true, 6)
	s.IO.Outln("2] Kiss her hand", true, 6)
	s.IO.Outln("3] Compliment and tip her", true, 6)
	s.IO.Outln("4] Brag about yourself", true, 6)
	s.IO.Outln("5] Wink seductively", true, 6)
	s.IO.Outln("6] Compliment her", true, 6)
	s.IO.Outln("7] Ask her to dance", true, 6)
	s.IO.Outln("8] Share a meal", true, 6)
	s.IO.Cr()

	num := s.IO.NumbersPrompt("Whats it gonna be Cutie?", 1, 8)

	s.IO.Cr()

	// Success chance: 18 + Flirt1%
	if mechanics.RandBetween(1, 100) < 18+s.Character.Flirt1 {
		s.IO.Outln("Corenne smiles at you as she serves you...", true, 5)
		s.IO.Cr()
		s.IO.Outln("Success!", true, 5)
		s.Character.Flirt1++
	} else {
		// Failure - show the text for the chosen flirt
		if num >= 1 && num <= 8 {
			s.IO.Outln(flirtFailTexts[num], true, 2)
		}

		// Failure penalties
		switch num {
		case 3, 4, 8:
			s.Character.Honor--
		case 6:
			dmg := mechanics.RandBetween(1, s.Character.HitPoints/4)
			s.Character.HitPoints -= dmg
		}

		s.Character.Flirt1 -= 2
		if s.Character.Flirt1 <= 0 {
			s.Character.Flirt1 = 1
		}
	}

	// Flip sign (cooldown - can only flirt once per visit)
	s.Character.Flirt1 = -s.Character.Flirt1

	s.IO.Cr()
	s.SetNextState("tavern_prompt")
}

// ViewGuildsState shows guild rankings.
type ViewGuildsState struct{}

func (ViewGuildsState) ID() game.StateID { return "view_guilds" }

func (ViewGuildsState) Enter(s *game.Session) {
	chars, err := s.Store.ListCharacters()
	if err != nil || len(chars) == 0 {
		s.IO.Outln("No adventurers have registered yet.", true, 1)
		s.SetNextState("tavern_prompt")
		return
	}

	s.IO.Cr()
	s.IO.Outln("=--=-- Guild Rankings ---=--=", true, 3)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf(" %-20s %-10s %5s %5s %5s %8s", "Name", "Class", "F.Lvl", "T.Lvl", "M.Lvl", "XP"), true, 4)
	s.IO.Outln(" "+fmt.Sprintf("%s", "------------------------------------------------------------"), true, 1)

	for _, c := range chars {
		s.IO.Outln(fmt.Sprintf(" %-20s %-10s %5d %5d %5d %8d",
			c.Name, c.CharClass, c.FighterLvl, c.ThiefLvl, c.MageLvl, c.TotalExperience), true, 1)
	}

	s.IO.Cr()
	s.IO.PausePrompt("--press a key--")
	s.SetNextState("tavern_prompt")
}
