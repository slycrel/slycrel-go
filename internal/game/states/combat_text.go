package states

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// --- Initiative text ---

var initiativeTextsPlayerFirst = []string{
	"You close with the vile %s, and snarl in defiance.\nThe %s quakes in fear as you ready your %s and square off...",
	"The terrifying %s quickly closes as you begin laughing.\nThis is going to be as easy as flirting with Corenne...",
	"A strange feeling comes over you as you begin the melee.\nA curse of rage is thrown your way from the %s in a guttural tongue...",
	"The %s leaps at you, inhuman appendages flailing.\nYou bring up your arms in defense as %s crashes into you. The battle begins!",
	"You ready your %s and prepare for the oncoming fray.\nThe surrounding terrain goes silent as you face off with the %s.",
	"The %s looks surprised that you do not quake in fear, and\nstarts waving its %s menacingly, attempting to ward you off.\nYou gleefully close for the kill!",
	"As you close with the %s, you cannot help but\ncompare it to your sixth grade teacher, Grizelda. As\nthe fight begins, you resolve to rid the world of both hideous beings...",
}

var initiativeTextsMonsterFirst = []string{
	"You cannot flee. The %s slashes at your back,\ncutting a painful gash in your left shoulder. You turn,\nand ready your %s, convinced that you are about to die...",
	"You bring up your %s, assured that you can\ndefeat the %s. As silence descends on the surrounding\narea, you start to wonder where the %s has gone to.\nThe sudden slice in your back whips you around, the fight already started...",
	"You bring up your %s, assured that you can\ndefeat the %s. As silence descends, you begin to\nwonder where that %s has gone to. The sudden\nslice in your back flips you around, the fiend already\npreparing for another strike...",
}

// SetupTextCombatState initializes text combat.
// Original: ST_SetupTextCombat
type SetupTextCombatState struct{}

func (SetupTextCombatState) ID() game.StateID { return "setup_text_combat" }

func (SetupTextCombatState) Enter(s *game.Session) {
	s.IO.ClearScreen()
	s.IO.Cr()

	// Init combat loop values (from original)
	s.UserAttackLoop = s.Character.Speed / 10
	if s.UserAttackLoop <= 0 {
		s.UserAttackLoop = 1
	}
	s.MonAttackLoop = s.Monster.Movement / 10
	if s.MonAttackLoop <= 0 {
		s.MonAttackLoop = 1
	}
	s.TextCombatLoop = 15
	s.Hit = model.HitNone
	s.UserMove = model.CombatAttack
	s.MonMove = model.CombatAttack
	s.LastDisposition = 0

	// Initiative roll
	gotInit := mechanics.RollInitiative(s.Character.Movement, s.Monster.Movement)
	showInitiativeText(s, gotInit)

	s.SetNextState("text_combat_loop")
}

func showInitiativeText(s *game.Session, playerFirst bool) {
	monName := s.Monster.Name
	weapName := s.Character.Weapons[0].Name
	if weapName == "" {
		weapName = "fists"
	}
	monWeap := s.Monster.AttackStr1

	if playerFirst {
		idx := rand.Intn(len(initiativeTextsPlayerFirst))
		template := initiativeTextsPlayerFirst[idx]
		var text string
		switch idx {
		case 0:
			text = fmt.Sprintf(template, monName, monName, weapName)
		case 1:
			text = fmt.Sprintf(template, monName)
		case 2:
			text = fmt.Sprintf(template, monName)
		case 3:
			text = fmt.Sprintf(template, monName, monName)
		case 4:
			text = fmt.Sprintf(template, weapName, monName)
		case 5:
			text = fmt.Sprintf(template, monName, monWeap)
		case 6:
			text = fmt.Sprintf(template, monName)
		}
		s.IO.Outln(text, true, 0)
	} else {
		idx := rand.Intn(len(initiativeTextsMonsterFirst))
		template := initiativeTextsMonsterFirst[idx]
		var text string
		switch idx {
		case 0:
			text = fmt.Sprintf(template, monName, weapName)
		case 1:
			text = fmt.Sprintf(template, weapName, monName, monName)
		case 2:
			text = fmt.Sprintf(template, weapName, monName, monName)
		}
		s.IO.Outln(text, true, 0)

		// Monster gets a free hit when it has initiative
		damage := mechanics.DoDamageMonster(s.Character, s.Monster)
		if damage > 0 {
			s.Character.HitPoints -= damage
			s.IO.Cr()
			s.IO.Outln(fmt.Sprintf("You are hit for %d damage.", damage), true, 6)
		}
	}
	s.IO.Cr()
}

// TextCombatLoopState is the main combat loop dispatcher.
// Original: ST_TextCombatLoop
type TextCombatLoopState struct{}

func (TextCombatLoopState) ID() game.StateID { return "text_combat_loop" }

func (TextCombatLoopState) Enter(s *game.Session) {
	// Check end conditions first
	if s.Character.HitPoints <= 0 {
		s.SetNextState("user_killed")
		return
	}
	if s.Monster.HitPoints <= 0 {
		s.SetNextState("user_victorious")
		return
	}
	if s.Hit == model.HitUserRan {
		s.SetNextState("user_escapes")
		return
	}
	if s.Hit == model.HitOpponentRan {
		s.SetNextState("monster_runs")
		return
	}

	// Determine whose turn it is based on combat loop
	if s.UserAttackLoop >= s.TextCombatLoop && s.Hit == model.HitNone {
		s.SetNextState("user_attack_stage")
	} else if s.MonAttackLoop >= s.TextCombatLoop &&
		s.Hit != model.HitOpponent &&
		s.Hit != model.HitOpponentOnly &&
		s.Hit != model.HitBoth &&
		s.Hit != model.HitNeither &&
		s.Hit != model.HitUserOnly &&
		s.Hit != model.HitOpponentRan {
		s.SetNextState("opponent_attack_stage")
	} else {
		s.TextCombatLoop--
		s.Hit = model.HitNone
		if s.TextCombatLoop <= 0 {
			s.TextCombatLoop = 27
		}
		// Loop back
		s.SetNextState("text_combat_loop")
	}
}

// UserAttackStageState shows disposition text and prompts for action.
// Original: ST_UserAttackStage + ST_CalcAdvanceMoves
type UserAttackStageState struct{}

func (UserAttackStageState) ID() game.StateID { return "user_attack_stage" }

func (UserAttackStageState) Enter(s *game.Session) {
	// Show disposition text if it changed
	showDispositionText(s)

	s.IO.Cr()
	choice := s.IO.LettersPrompt("A]ttack, B]lock, P]arry, D]odge, S]pecial, R]un --", "ABPDSR", 1, true, true)

	// Calculate advance moves
	advMoves := mechanics.CalcAdvanceMoves(s.UserAttackLoop, s.MonAttackLoop, s.TextCombatLoop)

	// Process user's choice
	processUserMove(s, choice, advMoves)

	s.SetNextState("text_combat_loop")
}

func processUserMove(s *game.Session, move string, advMoves int) {
	switch move {
	case "A":
		s.UserMove = model.CombatAttack
		s.IO.Outln("You attack the sucker!", true, 1)
		s.IO.Cr()

		damage := mechanics.DoDamageUser(s.Character, s.Monster)
		damage = mechanics.ApplyUserAttackModifier(damage, s.MonMove)

		// Dodge check
		if s.MonMove == model.CombatDodge {
			if mechanics.RandBetween(1, s.Character.Dexterity) < s.Monster.Movement/3 {
				damage = 0
			}
			if damage == 0 {
				s.IO.Outln("The monster laughs as it dodges your attack.", true, 5)
				s.IO.Cr()
			}
		}

		totalDamage := damage

		if totalDamage > 0 {
			s.IO.Outln(fmt.Sprintf("You hit the monster for %d damage.", totalDamage), true, 2)
			s.Hit = model.HitTheUser
		} else {
			s.IO.Cr()
			s.IO.Outln("You miss!", true, 6)
			s.IO.Cr()
		}

		// Advance moves (extra attacks from speed advantage)
		for i := 0; i < advMoves; i++ {
			extraDmg := mechanics.DoDamageUser(s.Character, s.Monster)
			totalDamage += extraDmg
			s.IO.Outln(fmt.Sprintf("Again, You hit the vile fiend for %d Damage!", extraDmg), true, 1)
			s.Hit = model.HitTheUser
		}

		s.Monster.HitPoints -= totalDamage
		s.Hit = model.HitTheUser

	case "B":
		s.UserMove = model.CombatBlock
		s.Hit = model.HitUserMiss
		s.IO.Outln("You raise your guard, ready to block.", true, 3)

	case "P":
		s.UserMove = model.CombatParry
		s.Hit = model.HitTheUser
		damage := mechanics.DoDamageUser(s.Character, s.Monster) / 3
		if damage > 0 {
			s.IO.Outln(fmt.Sprintf("You parry, slicing the monster for %d damage.", damage), true, 1)
		} else {
			s.IO.Outln("You raise your weapon, ready to parry.", true, 3)
		}
		s.IO.Cr()
		s.Monster.HitPoints -= damage

	case "D":
		s.UserMove = model.CombatDodge
		s.Hit = model.HitUserMiss
		s.IO.Outln("You prepare to dodge...", true, 3)

	case "S":
		s.UserMove = model.CombatSpecial
		s.Hit = model.HitUserMiss
		s.IO.Cr()
		s.IO.Outln("Okay, you sit there and feel special.", true, 1)

	case "R":
		s.IO.Cr()
		s.IO.Outln("You attempt escape...", true, 2)
		s.IO.Cr()
		s.Hit = model.HitUserRan
		s.UserMove = model.CombatSpecial
	}

	s.IO.Cr()
}

// OpponentAttackStageState handles the monster's turn.
// Original: ST_OpponentAttackStage
type OpponentAttackStageState struct{}

func (OpponentAttackStageState) ID() game.StateID { return "opponent_attack_stage" }

func (OpponentAttackStageState) Enter(s *game.Session) {
	// Calculate monster's advance moves
	advMoves := mechanics.CalcAdvanceMoves(s.MonAttackLoop, s.UserAttackLoop, s.TextCombatLoop)

	// Pick monster action
	monAction, monRan := mechanics.RandMonsterAction(s.Monster.HitPoints, s.MonHitPoints)

	if monRan {
		s.Hit = model.HitOpponentRan
		s.SetNextState("text_combat_loop")
		return
	}

	switch monAction {
	case model.CombatAttack:
		s.MonMove = model.CombatAttack
		damage := mechanics.DoDamageMonster(s.Character, s.Monster)
		damage = mechanics.ApplyMonsterAttackModifier(damage, s.UserMove, s.Character.Dexterity, s.Monster.Movement)

		if s.UserMove == model.CombatDodge && damage == 0 {
			s.IO.Outln("You Dodge successfully!", true, 5)
			s.IO.Cr()
		}
		if s.UserMove == model.CombatSpecial && damage > 0 {
			s.IO.Outln("You feel VERY special now.", true, 2)
			s.IO.Cr()
		}

		totalDamage := damage

		if totalDamage > 0 {
			s.IO.Cr()
			s.IO.Outln("The monster attacks you!", true, 3)
			s.IO.Cr()
			s.IO.Outln(fmt.Sprintf("You are hit for %d damage. (Feel better yet?)", totalDamage), true, 1)
			s.IO.Cr()

			// Advance moves
			for i := 0; i < advMoves; i++ {
				extraDmg := mechanics.DoDamageMonster(s.Character, s.Monster)
				extraDmg = mechanics.ApplyMonsterAttackModifier(extraDmg, s.UserMove, s.Character.Dexterity, s.Monster.Movement)
				totalDamage += extraDmg
				s.IO.Outln(fmt.Sprintf("You are hit again for %d damage!", extraDmg), true, 6)
			}

			s.Character.HitPoints -= totalDamage

			// Update hit tracking
			if s.Hit == model.HitTheUser {
				s.Hit = model.HitBoth
			} else if s.Hit == model.HitUserMiss || s.Hit == model.HitNone {
				s.Hit = model.HitOpponentOnly
			}
		} else {
			s.IO.Outln("The monster misses you!", true, 3)
			if s.Hit == model.HitTheUser {
				s.Hit = model.HitUserOnly
			} else if s.Hit == model.HitUserMiss || s.Hit == model.HitNone {
				s.Hit = model.HitNeither
			}
		}

	case model.CombatBlock:
		s.IO.Outln("The monster keeps its distance, circling warily.", true, 1)
		s.IO.Cr()
		s.MonMove = model.CombatBlock
		if s.Hit == model.HitTheUser {
			s.Hit = model.HitUserOnly
		} else {
			s.Hit = model.HitNeither
		}

	case model.CombatParry:
		s.IO.Cr()
		s.IO.Outln("The monster darts in and out, you aren't quite sure if it's attacking or not.", true, 2)
		s.IO.Cr()
		s.MonMove = model.CombatParry
		if s.Hit == model.HitTheUser {
			s.Hit = model.HitUserOnly
		} else {
			s.Hit = model.HitOpponentOnly
		}

	case model.CombatDodge:
		s.IO.Cr()
		s.IO.Outln("The monster backs away, hissing at you, looking like a coiled spring, ready to snap...", true, 3)
		s.IO.Cr()
		s.MonMove = model.CombatDodge
		if s.Hit == model.HitTheUser {
			s.Hit = model.HitUserOnly
		} else {
			s.Hit = model.HitNeither
		}
	}

	s.SetNextState("text_combat_loop")
}

// UserKilledState handles player death.
// Original: ST_UserKilled
type UserKilledState struct{}

func (UserKilledState) ID() game.StateID { return "user_killed" }

func (UserKilledState) Enter(s *game.Session) {
	if s.Character.Location == model.TheArenaCombat {
		s.Character.Location = model.TheArenaMenu
		s.SetNextState("arena_challenge_lost")
		return
	}

	s.IO.Cr()
	s.IO.Outln("Tough luck. You're dead, buddy.", true, 4)
	s.IO.Outln("Seeya in Hero's Heaven!", true, 5)
	s.IO.Cr()

	// Death penalty: lose 1/3 of coins on hand
	s.Character.CoinsHand -= s.Character.CoinsHand / 3
	s.Character.Alive = false

	s.Store.SaveCharacter(s.Character)
	s.SetNextState("dead")
}

// UserVictoriousState handles winning a fight.
// Original: ST_UserVictorious
type UserVictoriousState struct{}

func (UserVictoriousState) ID() game.StateID { return "user_victorious" }

func (UserVictoriousState) Enter(s *game.Session) {
	if s.Character.Location == model.TheArenaCombat {
		s.Character.Location = model.TheArenaMenu
		s.SetNextState("arena_challenge_won")
		return
	}

	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("You have defeated the %s!!", s.Monster.Name), true, 3)
	s.IO.Cr()

	// Rewards
	coinReward := int64(rand.Intn(s.Monster.Coins + 1))
	xpReward := int64(s.Monster.Experience)

	s.Character.TotalExperience += xpReward
	s.Character.SpendingExperience += xpReward
	s.Character.CoinsHand += coinReward
	s.Character.FightsWon++

	s.IO.Outln(fmt.Sprintf("You receive %d Experience points.", xpReward), true, 5)
	s.IO.Outln(fmt.Sprintf("You receive %d Coins.", coinReward), true, 5)
	s.IO.Cr()

	// Show current HP
	s.IO.Outln(fmt.Sprintf("Hit Points: %d/%d", s.Character.HitPoints, s.Character.MaxHP), true, 1)
	s.IO.Cr()

	s.Character.Location = model.TheWildernessMenu
	s.IO.PausePrompt("-=Press A Key=-")
	s.SetNextState("wilderness")
}

// UserEscapesState handles the player running away.
// Original: ST_UserEscapes
type UserEscapesState struct{}

func (UserEscapesState) ID() game.StateID { return "user_escapes" }

func (UserEscapesState) Enter(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln("You try to escape...", true, 3)

	// Dramatic dots
	dots := ""
	for i := 0; i < 40; i++ {
		dots += "."
	}
	s.IO.Outln(dots, true, 2)
	s.IO.Cr()
	s.IO.Outln("You Succeed!", true, 6)
	s.IO.Cr()

	s.IO.PausePrompt("--press a key--")
	if s.Character.Location == model.TheArenaCombat {
		s.Character.Location = model.TheArenaMenu
		s.SetNextState("arena")
	} else {
		s.Character.Location = model.TheWildernessMenu
		s.SetNextState("wilderness")
	}
}

// MonsterRunsState handles the monster fleeing.
// Original: ST_MonsterRuns
type MonsterRunsState struct{}

func (MonsterRunsState) ID() game.StateID { return "monster_runs" }

func (MonsterRunsState) Enter(s *game.Session) {
	s.IO.Cr()
	s.IO.Outln("<blink blink>", true, 1)
	s.IO.Cr()
	s.IO.Outln("The monster ran away from you! What a coward.", true, 1)
	s.IO.Cr()

	s.IO.PausePrompt("-More-")
	if s.Character.Location == model.TheArenaCombat {
		s.Character.Location = model.TheArenaMenu
		s.SetNextState("arena")
	} else {
		s.Character.Location = model.TheWildernessMenu
		s.SetNextState("wilderness")
	}
}

// showDispositionText displays narrative text based on the HP status of both combatants.
func showDispositionText(s *game.Session) {
	disp := mechanics.GetDisposition(
		s.Character.HitPoints, s.Character.MaxHP,
		s.Monster.HitPoints, s.MonHitPoints,
	)

	dispInt := int(disp)
	if dispInt == s.LastDisposition {
		return // only show when disposition changes
	}
	s.LastDisposition = dispInt

	// Generate disposition-appropriate flavor text
	monName := s.Monster.Name
	userPct := 0
	if s.Character.MaxHP > 0 {
		userPct = int(math.Round(float64(s.Character.HitPoints) / float64(s.Character.MaxHP) * 100))
	}
	monPct := 0
	if s.MonHitPoints > 0 {
		monPct = int(math.Round(float64(s.Monster.HitPoints) / float64(s.MonHitPoints) * 100))
	}

	s.IO.Cr()
	switch disp {
	case mechanics.DispMgoodUgood:
		s.IO.Outln(fmt.Sprintf("Both you and the %s are in good shape. This could go either way...", monName), true, 1)
	case mechanics.DispMgoodUbad:
		s.IO.Outln(fmt.Sprintf("The %s looks barely scratched, while you can barely stand...", monName), true, 6)
	case mechanics.DispMbadUgood:
		s.IO.Outln(fmt.Sprintf("The %s is staggering! You can taste victory!", monName), true, 3)
	case mechanics.DispMokUok:
		s.IO.Outln(fmt.Sprintf("You and the %s trade blows evenly...", monName), true, 4)
	case mechanics.DispMokUgood:
		s.IO.Outln(fmt.Sprintf("You have the advantage! The %s is starting to weaken.", monName), true, 3)
	case mechanics.DispMokUbad:
		s.IO.Outln(fmt.Sprintf("Things aren't looking great... the %s presses its advantage.", monName), true, 6)
	case mechanics.DispMgoodUok:
		s.IO.Outln(fmt.Sprintf("The %s is still going strong. You need to press harder!", monName), true, 4)
	case mechanics.DispMbadUok:
		s.IO.Outln(fmt.Sprintf("The %s is weakening! Keep it up!", monName), true, 3)
	case mechanics.DispMbadUbad:
		s.IO.Outln(fmt.Sprintf("You and the %s are both barely standing... one more hit could end it!", monName), true, 6)
	}
	s.IO.Outln(fmt.Sprintf("  [You: %d%% | %s: %d%%]", userPct, monName, monPct), true, 1)
	s.IO.Cr()
}
