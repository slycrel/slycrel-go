package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// BankState shows the bank ANSI menu.
// Original: ST_Bank
type BankState struct{}

func (BankState) ID() game.StateID { return "bank" }

func (BankState) Enter(s *game.Session) {
	s.Character.Location = model.TheBank
	s.IO.ShowANSIFile("bank_menu")
	s.SetNextState("bank_prompt")
}

// BankPromptState shows balances and prompts for action.
// Original: ST_BankPrompt + ST_BankMenu
type BankPromptState struct{}

func (BankPromptState) ID() game.StateID { return "bank_prompt" }

func (BankPromptState) Enter(s *game.Session) {
	s.IO.Outln(fmt.Sprintf("You have %d Coins in Bank and %d Coins on Hand.",
		s.Character.CoinsBank, s.Character.CoinsHand), true, 4)
	s.IO.Outln("[D]eposit, [W]ithdraw, [Q]uit, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "DWQ?", 1, true, true)

	switch choice {
	case "D":
		s.SetNextState("bank_deposit")
	case "W":
		s.SetNextState("bank_withdraw")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "?":
		s.SetNextState("bank")
	default:
		s.SetNextState("bank_prompt")
	}
}

// BankDepositState handles depositing coins.
// Original: ST_GetDeposit + ST_SaveDeposit
type BankDepositState struct{}

func (BankDepositState) ID() game.StateID { return "bank_deposit" }

func (BankDepositState) Enter(s *game.Session) {
	if s.Character.CoinsHand <= 0 {
		s.IO.Outln("You have no coins to deposit!", true, 6)
		s.IO.Cr()
		s.SetNextState("bank_prompt")
		return
	}

	s.IO.Outln(fmt.Sprintf("[Max: %d, 0 = Abort]", s.Character.CoinsHand), true, 5)
	s.IO.Cr()
	amount := s.IO.NumbersPrompt("Deposit:", 0, int(s.Character.CoinsHand))

	if amount > 0 {
		s.Character.CoinsBank += int64(amount)
		s.Character.CoinsHand -= int64(amount)
		s.IO.Outln(fmt.Sprintf("You give the banker %d coins.", amount), true, 3)
		s.IO.Cr()
		s.Store.SaveCharacter(s.Character)
	}

	s.SetNextState("bank_prompt")
}

// BankWithdrawState handles withdrawing coins.
// Original: ST_GetWithdrawl + ST_SaveWithdrawl
type BankWithdrawState struct{}

func (BankWithdrawState) ID() game.StateID { return "bank_withdraw" }

func (BankWithdrawState) Enter(s *game.Session) {
	if s.Character.CoinsBank <= 0 {
		s.IO.Outln("You have no coins in the bank!", true, 6)
		s.IO.Cr()
		s.SetNextState("bank_prompt")
		return
	}

	s.IO.Outln(fmt.Sprintf("[Max: %d, 0 = Abort]", s.Character.CoinsBank), true, 5)
	s.IO.Cr()
	amount := s.IO.NumbersPrompt("Withdraw:", 0, int(s.Character.CoinsBank))

	if amount > 0 {
		s.Character.CoinsHand += int64(amount)
		s.Character.CoinsBank -= int64(amount)
		s.IO.Outln(fmt.Sprintf("The Banker hands you %d coins.", amount), true, 3)
		s.IO.Cr()
		s.Store.SaveCharacter(s.Character)
	}

	s.SetNextState("bank_prompt")
}
