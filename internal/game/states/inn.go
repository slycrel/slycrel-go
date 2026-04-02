package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/model"
)

// InnState shows the inn ANSI menu.
type InnState struct{}

func (InnState) ID() game.StateID { return "inn" }

func (InnState) Enter(s *game.Session) {
	s.Character.Location = model.TheInn
	s.IO.ShowANSIFile("inn_menu")
	s.SetNextState("inn_prompt")
}

// InnPromptState shows the inn menu.
type InnPromptState struct{}

func (InnPromptState) ID() game.StateID { return "inn_prompt" }

func (InnPromptState) Enter(s *game.Session) {
	s.IO.Outln("[L]ook around, [R]ent a Room, [T]alk to Innkeeper, [V]iew, [Q]uit, [*]Owner Menu, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "LRTQV?*", 1, true, true)

	switch choice {
	case "L":
		s.IO.Outln("You look around the lobby of the inn and find it to be", true, 2)
		s.IO.Outln("in a serious array of disorder. You wonder if you", true, 2)
		s.IO.Outln("should really rent a room here or not.", true, 2)
		s.IO.Cr()
		s.SetNextState("inn_prompt")
	case "R":
		s.SetNextState("rent_room")
	case "T":
		s.IO.Outln("You ask the innkeep how business has been. Quickly", true, 2)
		s.IO.Outln("you realize you should have never asked such an open", true, 2)
		s.IO.Outln("ended question as the inkeep rambles on and on......", true, 2)
		s.IO.Outln("and on......", true, 2)
		s.IO.Cr()
		s.SetNextState("inn_prompt")
	case "V":
		s.IO.Outln("You look around the inn's lobby and quickly find", true, 2)
		s.IO.Outln("there really isn't anything to view.", true, 2)
		s.IO.Cr()
		s.SetNextState("inn_prompt")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "*":
		s.SetNextState("inn_owner_menu")
	case "?":
		s.SetNextState("inn")
	default:
		s.SetNextState("inn_prompt")
	}
}

// RentRoomState handles room rental.
type RentRoomState struct{}

func (RentRoomState) ID() game.StateID { return "rent_room" }

func (RentRoomState) Enter(s *game.Session) {
	inn, err := s.Store.LoadInn()
	if err != nil {
		s.IO.Outln("The inn seems to be closed...", true, 6)
		s.SetNextState("inn_prompt")
		return
	}

	// Check if player already has a room
	for _, room := range inn.Rooms {
		if room.Who == s.Character.Name && room.DaysLeft > 0 {
			s.IO.Outln("Sorry, you can only have one room.", true, 3)
			s.SetNextState("inn_prompt")
			return
		}
	}

	// Find empty room
	emptySlot := -1
	for i, room := range inn.Rooms {
		if room.Who == "" || room.DaysLeft <= 0 {
			emptySlot = i
			break
		}
	}
	if emptySlot == -1 {
		s.IO.Outln("Sorry There are no more rooms Avbl. for Rent", true, 3)
		s.SetNextState("inn_prompt")
		return
	}

	if !s.IO.YesNoQuestion("Would you like to Rent a room?") {
		s.SetNextState("inn_prompt")
		return
	}

	days := s.IO.NumbersPrompt("For how many days? [0-14]:", 0, 14)
	if days == 0 {
		s.IO.Outln("FINE! Don't Rent a room, go sleep in the streets!", true, 3)
		s.SetNextState("inn_prompt")
		return
	}

	rate := inn.CurRate
	if rate < 1 {
		rate = 5
	}
	totalCost := int64(days * rate)

	if s.Character.CoinsHand < totalCost {
		s.IO.Outln("I'm not giving the room away, come back when you have enough coins!", true, 3)
		s.IO.Outln(fmt.Sprintf("Note, it costs %d coins per day to rent a room.", rate), true, 1)
		s.SetNextState("inn_prompt")
		return
	}

	lockRating := s.IO.NumbersPrompt("What Security Rating would you like on your lock? [0-10]:", 0, 10)
	lockCost := int64(lockRating * 2)

	if lockRating == 0 {
		s.IO.Outln("\"Ok, don't blame me if Corenne slips in late at night,\" grins the Innkeeper.", true, 3)
	} else if s.Character.CoinsHand < totalCost+lockCost {
		s.IO.Outln("I'm not giving the Lock away!", true, 3)
		s.IO.Outln("Note, it costs twice the lock rating for the lock.", true, 1)
		lockCost = 0
		lockRating = 0
	}

	// Display receipt
	s.IO.Outln(fmt.Sprintf("Item:              Days:                  Cost:"), true, 2)
	s.IO.Outln(fmt.Sprintf(" Room               %-23d$%d", days, totalCost), true, 1)
	if lockCost > 0 {
		s.IO.Outln(fmt.Sprintf(" Lock               %-23d$%d", 1, lockCost), true, 1)
	}
	s.IO.Outln(fmt.Sprintf("Total:                                     $%d", totalCost+lockCost), true, 3)

	s.Character.CoinsHand -= totalCost + lockCost

	// Add to room
	inn.Rooms[emptySlot] = model.RoomRec{
		Who:      s.Character.Name,
		DaysLeft: days,
		Lock:     lockRating,
	}
	inn.Safe += totalCost
	s.Store.SaveInn(inn)

	s.IO.Cr()
	s.SetNextState("inn_prompt")
}

// InnOwnerMenuState shows the innkeeper's management menu.
type InnOwnerMenuState struct{}

func (InnOwnerMenuState) ID() game.StateID { return "inn_owner_menu" }

func (InnOwnerMenuState) Enter(s *game.Session) {
	inn, err := s.Store.LoadInn()
	if err != nil {
		s.SetNextState("inn_prompt")
		return
	}

	// Check if player is the owner
	if inn.Owner != "" && inn.Owner != s.Character.Name {
		s.IO.Outln("You are not the innkeeper!", true, 6)
		s.SetNextState("inn_prompt")
		return
	}

	// If no owner, claim it
	if inn.Owner == "" {
		inn.Owner = s.Character.Name
		s.Store.SaveInn(inn)
		s.IO.Outln("You are now the innkeeper!", true, 3)
		s.IO.Cr()
	}

	s.IO.Cr()
	s.IO.Outln("=--=-- Innkeeper's Menu ---=--=", true, 3)
	s.IO.Cr()
	s.IO.Outln("1) Change Daily Rate", true, 1)
	s.IO.Outln("2) List Rooms", true, 1)
	s.IO.Outln("3) Kick Out Player", true, 1)
	s.IO.Outln("4) View Inn Stats", true, 1)
	s.IO.Outln("5) Close/Open Inn", true, 1)
	s.IO.Outln("6) Transfer Ownership", true, 1)
	s.IO.Cr()

	num := s.IO.NumbersPrompt("[1-6, 0=Quit]:", 0, 6)

	switch num {
	case 0:
		s.SetNextState("inn_prompt")
	case 1:
		newRate := s.IO.NumbersPrompt("Daily Charge for a Room:", 1, 32000)
		inn.CurRate = newRate
		s.Store.SaveInn(inn)
		s.IO.Outln(fmt.Sprintf("Rate set to %d coins/day.", newRate), true, 3)
		s.SetNextState("inn_owner_menu")
	case 2:
		listRooms(s, inn)
		s.SetNextState("inn_owner_menu")
	case 3:
		listRooms(s, inn)
		roomNum := s.IO.NumbersPrompt("Which room to vacate? [1-10]:", 1, 10)
		inn.Rooms[roomNum-1] = model.RoomRec{}
		s.Store.SaveInn(inn)
		s.IO.Outln("Player removed.", true, 3)
		s.SetNextState("inn_owner_menu")
	case 4:
		s.IO.Cr()
		s.IO.Outln(fmt.Sprintf("  Owner: %s", inn.Owner), true, 4)
		s.IO.Outln(fmt.Sprintf("  Rate: %d coins/day", inn.CurRate), true, 1)
		s.IO.Outln(fmt.Sprintf("  Safe: %d coins", inn.Safe), true, 1)
		openStr := "Open"
		if !inn.Open {
			openStr = "Closed"
		}
		s.IO.Outln(fmt.Sprintf("  Status: %s", openStr), true, 1)
		s.IO.Cr()
		s.SetNextState("inn_owner_menu")
	case 5:
		inn.Open = !inn.Open
		s.Store.SaveInn(inn)
		if inn.Open {
			s.IO.Outln("The Inn is now OPEN.", true, 3)
		} else {
			s.IO.Outln("The Inn is now CLOSED.", true, 6)
		}
		s.SetNextState("inn_owner_menu")
	case 6:
		newOwner := s.IO.ReadLine("Who is to be the new Inn Owner?", 20)
		if newOwner != "" {
			inn.Owner = newOwner
			s.Store.SaveInn(inn)
			s.IO.Outln(fmt.Sprintf("Ownership transferred to %s.", newOwner), true, 3)
		}
		s.SetNextState("inn_owner_menu")
	default:
		s.SetNextState("inn_owner_menu")
	}
}

func listRooms(s *game.Session, inn *model.InnRec) {
	s.IO.Cr()
	s.IO.Outln("  Room  Occupant             Days Left  Lock", true, 4)
	s.IO.Outln("  ----  --------             ---------  ----", true, 1)
	for i, room := range inn.Rooms {
		who := "(empty)"
		if room.Who != "" && room.DaysLeft > 0 {
			who = room.Who
		}
		s.IO.Outln(fmt.Sprintf("  %2d    %-20s %5d      %d", i+1, who, room.DaysLeft, room.Lock), true, 1)
	}
	s.IO.Cr()
}
