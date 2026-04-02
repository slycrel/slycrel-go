package states

import (
	"fmt"
	"strings"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// ArenaState shows the arena ANSI menu.
type ArenaState struct{}

func (ArenaState) ID() game.StateID { return "arena" }

func (ArenaState) Enter(s *game.Session) {
	s.Character.Location = model.TheArenaMenu
	s.IO.ShowANSIFile("arena_menu")
	s.SetNextState("arena_prompt")
}

// ArenaPromptState shows the arena menu.
type ArenaPromptState struct{}

func (ArenaPromptState) ID() game.StateID { return "arena_prompt" }

func (ArenaPromptState) Enter(s *game.Session) {
	s.IO.Outln("[B]et, [C]hallenge, [G]ladiator Fight, [L]ist Players, [Q]uit, [V]iew Stats, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "BCGLQV?", 1, true, true)

	switch choice {
	case "B":
		s.SetNextState("arena_bet")
	case "C":
		s.SetNextState("arena_challenge")
	case "G":
		s.SetNextState("arena_gladiator")
	case "L":
		s.SetNextState("arena_list_players")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "V":
		s.SetReturnState("arena_prompt")
		s.SetNextState("view_character")
	case "?":
		s.SetNextState("arena")
	default:
		s.SetNextState("arena_prompt")
	}
}

// ArenaChallengeState handles direct PvP challenge (text combat).
type ArenaChallengeState struct{}

func (ArenaChallengeState) ID() game.StateID { return "arena_challenge" }

func (ArenaChallengeState) Enter(s *game.Session) {
	if s.Character.Spars < 1 {
		s.IO.Outln("Sorry Buddy, but you've fought All that you can Today.", true, 2)
		s.SetNextState("arena_prompt")
		return
	}

	name := s.IO.ReadLine("Player to Challenge:", 27)
	if name == "" {
		s.SetNextState("arena_prompt")
		return
	}

	// Find the opponent
	opponent, err := s.Store.FindCharacterByName(name)
	if err != nil {
		s.IO.Outln("Player Not Found", true, 6)
		s.SetNextState("arena_prompt")
		return
	}

	if !opponent.Alive {
		s.IO.Outln("That Player is Dead, you can't fight him!", true, 1)
		s.SetNextState("arena_prompt")
		return
	}

	if strings.EqualFold(opponent.Name, s.Character.Name) {
		s.IO.Outln("You can not challenge yourself. Fool!", true, 3)
		s.SetNextState("arena_prompt")
		return
	}

	// Convert opponent to monster and fight
	monster := mechanics.User2Monster(opponent)
	s.Monster = &monster
	s.MonHitPoints = monster.HitPoints
	s.CombatName = opponent.Name
	s.Character.Location = model.TheArenaCombat
	s.Character.Spars--
	s.Character.TotalFights++

	spd := s.Character.Speed
	if spd < 5 {
		spd = 5
	}
	s.Character.Movement = randBetween(spd-spd/5, spd+spd/5)

	s.IO.ClearScreen()
	s.IO.Outln(fmt.Sprintf("You face %s in the Arena!", opponent.Name), true, 6)
	s.IO.Cr()

	// Arena uses arena terrain maps
	s.CombatRegion = "arena"
	s.SetNextState("grid_combat_setup")
}

// ArenaGladiatorState sets up a gladiator fight (resolved at daily upkeep).
type ArenaGladiatorState struct{}

func (ArenaGladiatorState) ID() game.StateID { return "arena_gladiator" }

func (ArenaGladiatorState) Enter(s *game.Session) {
	if s.Character.Spars < 1 {
		s.IO.Outln("Sorry Buddy, but you've fought All that you can Today.", true, 2)
		s.SetNextState("arena_prompt")
		return
	}

	name := s.IO.ReadLine("Player to Challenge:", 27)
	if name == "" {
		s.SetNextState("arena_prompt")
		return
	}

	opponent, err := s.Store.FindCharacterByName(name)
	if err != nil {
		s.IO.Outln("Player Not Found", true, 6)
		s.SetNextState("arena_prompt")
		return
	}

	if !opponent.Alive {
		s.IO.Outln("That Player is Dead, you can't fight him!", true, 1)
		s.SetNextState("arena_prompt")
		return
	}

	if strings.EqualFold(opponent.Name, s.Character.Name) {
		s.IO.Outln("You can not challenge yourself. Fool!", true, 3)
		s.SetNextState("arena_prompt")
		return
	}

	if !s.IO.YesNoQuestion("Are you sure you want to go through with this? [Y/n]") {
		s.IO.Cr()
		s.IO.Outln("You regretfully decide not to go through with this.", true, 1)
		s.SetNextState("arena_prompt")
		return
	}

	// Calculate odds
	challengerVal := mechanics.ArenaAppraisal(s.Character)
	opponentVal := mechanics.ArenaAppraisal(opponent)
	odds := mechanics.CalcOdds(challengerVal, opponentVal)

	// Save gladiator fight
	fight := &model.GCombatPrefs{
		Challenger: s.Character.Name,
		Opponent:   opponent.Name,
		Odds:       odds,
	}
	s.Store.SaveGladiatorFight(fight)

	s.Character.Spars--

	s.IO.Outln(fmt.Sprintf("You have entered into the Gladiator competition against %s (%s)!",
		opponent.Name, opponent.BBSName), true, 1)
	s.IO.Outln(fmt.Sprintf("Odds (Challenger:Opponent) = %d:%d", odds[0], odds[1]), true, 4)
	s.IO.Cr()

	s.SetNextState("arena_prompt")
}

// ArenaBetState handles betting on gladiator fights.
type ArenaBetState struct{}

func (ArenaBetState) ID() game.StateID { return "arena_bet" }

func (ArenaBetState) Enter(s *game.Session) {
	fights, err := s.Store.LoadGladiatorFights()
	if err != nil || len(fights) == 0 {
		s.IO.Outln("Sorry, there are no matches to bet on at this time.", true, 3)
		s.IO.Cr()
		s.SetNextState("arena_prompt")
		return
	}

	// Display fights
	s.IO.Cr()
	for i, f := range fights {
		s.IO.Outln(fmt.Sprintf("%d) %s vs %s  (Odds %d:%d)",
			i+1, f.Challenger, f.Opponent, f.Odds[0], f.Odds[1]), true, 1)
	}
	s.IO.Cr()

	fightNum := s.IO.NumbersPrompt("Which fight? (0=Quit):", 0, len(fights))
	if fightNum == 0 {
		s.SetNextState("arena_prompt")
		return
	}

	fight := fights[fightNum-1]

	// Can't bet on your own fight
	if strings.EqualFold(fight.Challenger, s.Character.Name) ||
		strings.EqualFold(fight.Opponent, s.Character.Name) {
		s.IO.Outln("Sorry, you can't bet when you're already busy fighting.", true, 3)
		s.SetNextState("arena_prompt")
		return
	}

	choice := s.IO.LettersPrompt("Bet [F]or or [A]gainst the challenger? (Q=Quit)", "FAQ", 1, true, true)
	if choice == "Q" {
		s.SetNextState("arena_prompt")
		return
	}

	forChallenger := choice == "F"

	maxBet := 1000
	if s.Character.CoinsHand < int64(maxBet) {
		maxBet = int(s.Character.CoinsHand)
	}
	if maxBet <= 0 {
		s.IO.Outln("You don't have any coins to bet!", true, 6)
		s.SetNextState("arena_prompt")
		return
	}

	amount := s.IO.NumbersPrompt(fmt.Sprintf("How much do you wish to bet? [1..%d]:", maxBet), 1, maxBet)

	if int64(amount) > s.Character.CoinsHand {
		s.IO.Outln("Sorry, you don't have enough cash.", true, 1)
		s.IO.Cr()
		s.SetNextState("arena_prompt")
		return
	}

	s.Character.CoinsHand -= int64(amount)

	bet := &model.GCombatBet{
		FightNum:      fightNum,
		FileName:      s.Character.Name,
		Bet:           amount,
		Challenger:    fight.Challenger,
		Opponent:      fight.Opponent,
		ForChallenger: forChallenger,
	}
	s.Store.SaveBet(bet)

	s.IO.Outln("Okay, you now are in the pool!", true, 2)
	s.IO.Cr()
	s.IO.PausePrompt("-Any Key-")
	s.SetNextState("arena_prompt")
}

// ArenaListPlayersState shows all players.
type ArenaListPlayersState struct{}

func (ArenaListPlayersState) ID() game.StateID { return "arena_list_players" }

func (ArenaListPlayersState) Enter(s *game.Session) {
	chars, err := s.Store.ListCharacters()
	if err != nil || len(chars) == 0 {
		s.IO.Outln("No players found.", true, 1)
		s.SetNextState("arena_prompt")
		return
	}

	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf(" %-25s %8s %-6s %-10s", "Name", "Score", "Alive", "Class"), true, 4)
	s.IO.Outln(" "+fmt.Sprintf("%s", "------------------------------------------------------"), true, 1)

	for _, c := range chars {
		alive := "Alive"
		if !c.Alive {
			alive = "Dead"
		}
		s.IO.Outln(fmt.Sprintf(" %-25s %8d %-6s %-10s",
			c.Name, c.ScoreVal, alive, c.CharClass), true, 1)
	}

	s.IO.Cr()
	s.IO.PausePrompt("--press a key--")
	s.SetNextState("arena_prompt")
}

// ArenaChallengeWonState handles winning an arena fight.
type ArenaChallengeWonState struct{}

func (ArenaChallengeWonState) ID() game.StateID { return "arena_challenge_won" }

func (ArenaChallengeWonState) Enter(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("You have Defeated %s!!", s.Monster.Name), true, 3)

	// Rewards
	coinReward := int64(randBetween(1, s.Monster.Coins+1))
	xpReward := int64(s.Monster.Experience)
	s.Character.TotalExperience += xpReward
	s.Character.SpendingExperience += xpReward
	s.Character.CoinsHand += coinReward
	s.Character.FightsWon++

	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("You receive %d Experience points.", xpReward), true, 5)
	s.IO.Outln(fmt.Sprintf("You receive %d Coins.", coinReward), true, 5)
	s.IO.Cr()

	// Damage the loser
	loser, err := s.Store.FindCharacterByName(s.CombatName)
	if err == nil {
		loser.Alive = false
		loser.CoinsHand -= coinReward
		if loser.CoinsHand < 0 {
			loser.CoinsHand = 0
		}
		loser.TotalExperience -= 20
		loser.SpendingExperience -= 20
		s.Store.SaveCharacter(loser)
		s.Store.WriteNews(loser.BBSName,
			fmt.Sprintf("You have been slaughtered by %s!", s.Character.Name))
	}

	s.IO.PausePrompt("-=Press A Key=-")
	s.Character.Location = model.TheArenaMenu
	s.SetNextState("arena")
}

// ArenaChallengeLostState handles losing an arena fight.
type ArenaChallengeLostState struct{}

func (ArenaChallengeLostState) ID() game.StateID { return "arena_challenge_lost" }

func (ArenaChallengeLostState) Enter(s *game.Session) {
	s.IO.Outln("You're outta life buddy...", true, 6)
	s.IO.Cr()

	// Reward the winner
	winner, err := s.Store.FindCharacterByName(s.CombatName)
	if err == nil {
		coinsTaken := int64(randBetween(1, int(s.Character.CoinsHand/2)+1))
		xpGained := int64(float64(s.Character.TotalExperience) * 0.3)

		winner.CoinsHand += coinsTaken
		winner.TotalExperience += xpGained
		winner.SpendingExperience += xpGained
		winner.FightsWon++
		winner.TotalFights++
		s.Store.SaveCharacter(winner)
		s.Store.WriteNews(winner.BBSName,
			fmt.Sprintf("You have slaughtered %s!", s.Character.Name))

		s.Character.CoinsHand -= coinsTaken
	}

	s.Character.Alive = false
	s.Character.TotalFights++
	s.Store.SaveCharacter(s.Character)

	s.SetNextState("dead")
}
