package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// ArmoryState shows the armory ANSI menu.
// Original: ST_Armory
type ArmoryState struct{}

func (ArmoryState) ID() game.StateID { return "armory" }

func (ArmoryState) Enter(s *game.Session) {
	s.Character.Location = model.TheArmory
	s.IO.ShowANSIFile("armory_menu")
	s.SetNextState("armory_prompt")
}

// ArmoryPromptState shows the armory menu.
// Original: ST_ArmoryShortPrompt + ST_ArmoryMenu
type ArmoryPromptState struct{}

func (ArmoryPromptState) ID() game.StateID { return "armory_prompt" }

func (ArmoryPromptState) Enter(s *game.Session) {
	c := s.Character
	s.IO.Outln(fmt.Sprintf("  Coins on hand: %d", c.CoinsHand), true, 4)
	s.IO.Outln("[W]eapons, [A]rmor, [S]ell Weapon, [G] Sell Armor, [M]ain/Off Switch, [Q]uit, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "WASGMQ?", 1, true, true)

	switch choice {
	case "W":
		s.SetNextState("buy_weapon")
	case "A":
		s.SetNextState("buy_armor")
	case "S":
		s.SetNextState("sell_weapon")
	case "G":
		s.SetNextState("sell_armor")
	case "M":
		// Swap weapon slots
		s.IO.Cr()
		s.IO.Outln("You switch your weapons.", true, 1)
		s.IO.Cr()
		c.Weapons[0], c.Weapons[1] = c.Weapons[1], c.Weapons[0]
		s.SetNextState("armory_prompt")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "?":
		s.SetNextState("armory")
	default:
		s.SetNextState("armory_prompt")
	}
}

// BuyWeaponState lists and lets you buy weapons.
// Original: ST_OpenWeaponsFile + ST_ChooseWeapon + ST_QBuyWeapon + ST_QReplaceWeapon + ST_DisplayWeapons
type BuyWeaponState struct{}

func (BuyWeaponState) ID() game.StateID { return "buy_weapon" }

func (BuyWeaponState) Enter(s *game.Session) {
	weapons, err := s.Store.LoadWeapons()
	if err != nil || len(weapons) == 0 {
		s.IO.Outln("No weapons available!", true, 6)
		s.IO.PausePrompt("--press a key--")
		s.SetNextState("armory_prompt")
		return
	}

	s.IO.Cr()
	s.IO.Outln("=--=-- Available Weapons ---=--=", true, 3)
	s.IO.Cr()

	for i, w := range weapons {
		affordable := " "
		if int64(w.Cost) <= s.Character.CoinsHand {
			affordable = "*"
		}
		rangeStr := ""
		if w.Range > 0 {
			rangeStr = fmt.Sprintf(" Range:%d", w.Range)
		}
		s.IO.Outln(fmt.Sprintf(" %s%2d) %-18s Strike:%-3d%s  Cost:%-4d",
			affordable, i+1, w.Name, w.Strike, rangeStr, w.Cost), true, 1)
	}
	s.IO.Cr()
	s.IO.Outln("  (* = you can afford)", true, 4)
	s.IO.Cr()

	num := s.IO.NumbersPrompt("Choose Your Weapon (0=Quit):", 0, len(weapons))
	if num == 0 {
		s.SetNextState("armory_prompt")
		return
	}

	weapon := weapons[num-1]

	if !s.IO.YesNoQuestion(fmt.Sprintf("Buy %s for %d coins? [Y/n]", weapon.Name, weapon.Cost)) {
		s.IO.Outln("FINE, I didn't want your stupid weapon anyway...", true, 1)
		s.IO.Cr()
		s.SetNextState("armory_prompt")
		return
	}

	if s.Character.CoinsHand < int64(weapon.Cost) {
		s.IO.Cr()
		s.IO.Outln("Sorry, I'm not giving these away. Come back when you get more money.", true, 6)
		s.IO.Cr()
		s.IO.PausePrompt("--More--")
		s.SetNextState("armory_prompt")
		return
	}

	// Show current weapons and ask which slot
	s.IO.Cr()
	w1name := s.Character.Weapons[0].Name
	if w1name == "" {
		w1name = "Empty"
	}
	w2name := s.Character.Weapons[1].Name
	if w2name == "" {
		w2name = "Empty"
	}
	s.IO.Outln(fmt.Sprintf("  Weapon 1: %s", w1name), true, 1)
	s.IO.Outln(fmt.Sprintf("  Weapon 2: %s", w2name), true, 1)
	s.IO.Cr()

	slot := s.IO.NumbersPrompt("Which weapon slot to replace? (1 or 2):", 1, 2)

	s.Character.Weapons[slot-1] = weapon
	s.Character.CoinsHand -= int64(weapon.Cost)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("Okay, you buy the spiffy %s.", weapon.Name), true, 3)
	s.IO.PausePrompt("--More--")
	s.SetNextState("armory_prompt")
}

// BuyArmorState lists and lets you buy armor.
// Original: ST_OpenArmorFile + ST_ChooseArmor + ST_QBuyArmor + ST_SaveNewArmor
type BuyArmorState struct{}

func (BuyArmorState) ID() game.StateID { return "buy_armor" }

func (BuyArmorState) Enter(s *game.Session) {
	armors, err := s.Store.LoadArmor()
	if err != nil || len(armors) == 0 {
		s.IO.Outln("No armor available!", true, 6)
		s.IO.PausePrompt("--press a key--")
		s.SetNextState("armory_prompt")
		return
	}

	s.IO.Cr()
	s.IO.Outln("=--=-- Available Armor ---=--=", true, 3)
	s.IO.Cr()

	classNames := []string{"Body", "Shield", "Full Body"}
	for i, a := range armors {
		affordable := " "
		if int64(a.Cost) <= s.Character.CoinsHand {
			affordable = "*"
		}
		cn := "Body"
		if int(a.Class) < len(classNames) {
			cn = classNames[a.Class]
		}
		s.IO.Outln(fmt.Sprintf(" %s%2d) %-18s Def:%-3d Type:%-9s Cost:%-4d",
			affordable, i+1, a.Name, a.Defense, cn, a.Cost), true, 1)
	}
	s.IO.Cr()
	s.IO.Outln("  (* = you can afford)", true, 4)
	s.IO.Cr()

	num := s.IO.NumbersPrompt("Choose Your Armor (0=Quit):", 0, len(armors))
	if num == 0 {
		s.SetNextState("armory_prompt")
		return
	}

	armor := armors[num-1]

	if !s.IO.YesNoQuestion(fmt.Sprintf("Buy %s for %d coins? [Y/n]", armor.Name, armor.Cost)) {
		s.IO.Outln("Okay, your loss. Have fun dying.", true, 1)
		s.IO.Cr()
		s.SetNextState("armory_prompt")
		return
	}

	if s.Character.CoinsHand < int64(armor.Cost) {
		s.IO.Cr()
		s.IO.Outln("Sorry, I'm not giving these away. Come back when you get more money.", true, 6)
		s.IO.Cr()
		s.SetNextState("armory_prompt")
		return
	}

	// Replace body armor (slot 0)
	s.Character.Armor[0] = armor
	s.Character.CoinsHand -= int64(armor.Cost)
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("Okay, you buy the spiffy %s.", armor.Name), true, 3)
	s.IO.PausePrompt("--More--")
	s.SetNextState("armory_prompt")
}

// SellWeaponState sells the primary weapon.
// Original: ST_QSellWeapon + ST_SaveWeaponSale
type SellWeaponState struct{}

func (SellWeaponState) ID() game.StateID { return "sell_weapon" }

func (SellWeaponState) Enter(s *game.Session) {
	w := s.Character.Weapons[0]
	sellPrice := w.Cost / 2

	if w.Name == "" || w.Name == "Hands" {
		s.IO.Cr()
		s.IO.Outln("You can't sell your hands. Idiot...", true, 6)
		s.IO.Cr()
		s.IO.PausePrompt("-More-")
		s.SetNextState("armory_prompt")
		return
	}

	if w.Cost <= 0 {
		s.IO.Cr()
		s.IO.Outln("Sorry, I don't buy junk.", true, 6)
		s.IO.Cr()
		s.IO.PausePrompt("-More-")
		s.SetNextState("armory_prompt")
		return
	}

	if !s.IO.YesNoQuestion(fmt.Sprintf("Sell your %s for %d coins? [Y/n]", w.Name, sellPrice)) {
		s.IO.Cr()
		s.IO.Outln("Okay, you jerk, I didn't want it anyway.", true, 3)
		s.IO.Cr()
		s.IO.PausePrompt("-More-")
		s.SetNextState("armory_prompt")
		return
	}

	s.Character.CoinsHand += int64(sellPrice)
	s.Character.Weapons[0] = model.Weapon{Name: "Hands", Strike: 1, ActionStr: "punch"}
	s.IO.Cr()
	s.IO.Outln("Okay, thanks. Come back anytime.", true, 1)
	s.IO.Cr()
	s.IO.PausePrompt("-More-")
	s.SetNextState("armory_prompt")
}

// SellArmorState sells the body armor.
// Original: ST_QSellArmor + ST_SaveArmorSale
type SellArmorState struct{}

func (SellArmorState) ID() game.StateID { return "sell_armor" }

func (SellArmorState) Enter(s *game.Session) {
	a := s.Character.Armor[0]
	sellPrice := a.Cost / 2

	if a.Name == "" || a.Cost <= 0 {
		s.IO.Cr()
		s.IO.Outln("Sorry, I don't buy junk.", true, 6)
		s.IO.Cr()
		s.IO.PausePrompt("-More-")
		s.SetNextState("armory_prompt")
		return
	}

	if !s.IO.YesNoQuestion(fmt.Sprintf("Sell your %s for %d coins? [Y/n]", a.Name, sellPrice)) {
		s.IO.Cr()
		s.IO.Outln("Okay, you jerk, I didn't want it anyway.", true, 3)
		s.IO.Cr()
		s.IO.PausePrompt("-More-")
		s.SetNextState("armory_prompt")
		return
	}

	s.Character.CoinsHand += int64(sellPrice)
	s.Character.Armor[0] = model.Armor{}
	s.IO.Cr()
	s.IO.Outln("Okay, thanks. Come back anytime.", true, 1)
	s.IO.Cr()
	s.IO.PausePrompt("-More-")
	s.SetNextState("armory_prompt")
}
