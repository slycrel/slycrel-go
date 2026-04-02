package states

import (
	"fmt"
	"math"
	"strings"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// GridCombatSetupState loads terrain and places combatants.
// Replaces the old SetupCombatState for when terrain is available.
type GridCombatSetupState struct{}

func (GridCombatSetupState) ID() game.StateID { return "grid_combat_setup" }

func (GridCombatSetupState) Enter(s *game.Session) {
	// Pick a terrain map
	mapName := s.CombatRegion + "-1"
	n := randBetween(1, 3)
	mapName = fmt.Sprintf("%s-%d", s.CombatRegion, n)

	terrain, err := s.Store.LoadTerrain(mapName)
	if err != nil {
		// Fall back to text combat if no terrain available
		s.SetNextState("setup_text_combat")
		return
	}
	s.Terrain = terrain

	// Place player and monster randomly (avoid water/boulders, avoid same spot)
	for attempts := 0; attempts < 100; attempts++ {
		s.UserR = randBetween(0, model.GridRows-1)
		s.UserC = randBetween(0, model.GridCols-1)
		cost := mechanics.NumMovePointsAt(s.UserR, s.UserC, terrain)
		if cost < 998 {
			break
		}
	}
	for attempts := 0; attempts < 100; attempts++ {
		s.MonsR = randBetween(0, model.GridRows-1)
		s.MonsC = randBetween(0, model.GridCols-1)
		cost := mechanics.NumMovePointsAt(s.MonsR, s.MonsC, terrain)
		if cost < 998 && !(s.MonsR == s.UserR && s.MonsC == s.UserC) {
			break
		}
	}

	s.TextOutln = 0
	drawGridScreen(s)
	gridTextOut(s, fmt.Sprintf("You Encounter a %s", s.Monster.Name))

	s.SetNextState("grid_combat_prompt")
}

// GridCombatPromptState handles the main grid combat input loop.
// Original: ST_Combat + ST_MovePlayer
type GridCombatPromptState struct{}

func (GridCombatPromptState) ID() game.StateID { return "grid_combat_prompt" }

func (GridCombatPromptState) Enter(s *game.Session) {
	// Check death
	if s.Character.HitPoints <= 0 || !s.Character.Alive {
		s.SetNextState("user_killed")
		return
	}
	if s.Monster.HitPoints <= 0 {
		s.IO.ClearScreen()
		s.SetNextState("user_victorious")
		return
	}

	// Check if player can move
	if !mechanics.CanMoveAnywhere(s.UserR, s.UserC, s.Character.Movement, s.Terrain) {
		// Player out of movement - monster's turn
		s.SetNextState("grid_monster_move")
		return
	}

	// Update movement display
	updateMovementDisplay(s)

	choice := s.IO.LettersPrompt("", "AIJKLFRPS*", 1, true, false)

	switch choice {
	case "I": // up
		moveGridPlayer(s, -1, 0)
	case "K": // down
		moveGridPlayer(s, 1, 0)
	case "J": // left
		moveGridPlayer(s, 0, -1)
	case "L": // right
		moveGridPlayer(s, 0, 1)
	case "A": // attack (try to engage text combat)
		if s.Character.Movement < 3 {
			gridTextOut(s, "Not 'nuff Movement Pts. Left!")
		} else {
			s.Character.Movement -= 3
			if randBetween(1, 50) < 8 {
				s.SetNextState("setup_text_combat")
				return
			}
			gridTextOut(s, "You Attempt to Attack the Creature, & you fail.")
		}
	case "F": // fire/shoot ranged weapon
		if s.Character.Movement < 3 {
			gridTextOut(s, "Not 'nuff Movement Pts. Left!")
		} else {
			doGridShoot(s)
		}
	case "R": // run
		if s.Character.Location == model.TheArenaCombat {
			s.Character.Location = model.TheArenaMenu
			s.SetNextState("arena")
		} else {
			s.Character.Location = model.TheWildernessMenu
			s.SetNextState("wilderness")
		}
		return
	case "P": // pass
		gridTextOut(s, "You pass your remaining moves")
		s.SetNextState("grid_monster_move")
		return
	case "S": // special (pass with flair)
		gridTextOut(s, "You feel special while passing your moves")
		s.SetNextState("grid_monster_move")
		return
	case "*": // redraw
		drawGridScreen(s)
	}

	// Check collision - if player on monster, enter text combat
	if s.UserR == s.MonsR && s.UserC == s.MonsC {
		s.SetNextState("setup_text_combat")
		return
	}

	s.SetNextState("grid_combat_prompt")
}

// GridMonsterMoveState handles the monster's turn on the grid.
// Original: ST_MoveMonster
type GridMonsterMoveState struct{}

func (GridMonsterMoveState) ID() game.StateID { return "grid_monster_move" }

func (GridMonsterMoveState) Enter(s *game.Session) {
	// Monster pathfinds toward player
	route, found := mechanics.FindRoute(s.MonsR, s.MonsC, s.UserR, s.UserC, s.Terrain)

	if found && len(route) > 0 {
		moveLeft := s.Monster.Movement

		for _, dir := range route {
			newR, newC := s.MonsR, s.MonsC
			switch dir {
			case 'U':
				newR--
			case 'D':
				newR++
			case 'L':
				newC--
			case 'R':
				newC++
			}

			if newR < 0 || newR >= model.GridRows || newC < 0 || newC >= model.GridCols {
				continue
			}

			cost := mechanics.NumMovePointsAt(newR, newC, s.Terrain)
			if cost > moveLeft {
				break
			}

			// Erase old monster position
			drawTerrainCell(s, s.MonsR, s.MonsC)

			s.MonsR = newR
			s.MonsC = newC
			moveLeft -= cost

			// 20% chance to shoot
			if randBetween(1, 5) == 1 && moveLeft >= 3 {
				doGridMonsterShoot(s)
				moveLeft -= 3
			}

			if moveLeft < 1 {
				break
			}
		}

		// Draw monster at new position
		drawMonsterOnGrid(s)
	}

	// Reset player movement for next turn
	spd := s.Character.Speed
	if spd < 5 {
		spd = 5
	}
	s.Character.Movement = randBetween(spd-spd/5, spd+spd/5)

	// Check collision
	if s.UserR == s.MonsR && s.UserC == s.MonsC {
		s.SetNextState("setup_text_combat")
		return
	}

	s.SetNextState("grid_combat_prompt")
}

// --- Grid rendering helpers ---

func drawGridScreen(s *game.Session) {
	s.IO.ClearScreen()
	terrain := s.Terrain

	// Draw border top
	s.IO.ANSICode("1;1H")
	s.IO.ANSICode("1;37m+" + strings.Repeat("-", model.GridCols) + "+")

	// Draw terrain grid
	for r := 0; r < model.GridRows; r++ {
		s.IO.ANSICode(fmt.Sprintf("%d;1H", r+2))
		s.IO.ANSICode("1;37m|")
		for c := 0; c < model.GridCols; c++ {
			td := mechanics.GetTerrainDisplay(terrain.Cells[r][c])
			s.IO.ANSICode(td.ANSIColor + td.Char)
		}
		s.IO.ANSICode("1;37m|")
	}

	// Draw border bottom
	s.IO.ANSICode(fmt.Sprintf("%d;1H", model.GridRows+2))
	s.IO.ANSICode("1;37m+" + strings.Repeat("-", model.GridCols) + "+")

	// Draw player and monster
	drawPlayerOnGrid(s)
	drawMonsterOnGrid(s)

	// Draw stats panel
	drawGridStats(s)

	// Draw text area separator
	s.IO.ANSICode(fmt.Sprintf("%d;1H", model.GridRows+3))
	s.IO.ANSICode("0;36m" + strings.Repeat("-", 50))

	// Position cursor for input
	s.IO.ANSICode(fmt.Sprintf("%d;1H", model.GridRows+9))
	s.IO.ANSICode("0;36m[I/J/K/L]Move [A]ttack [F]ire [R]un [P]ass [*]Redraw")
}

func drawPlayerOnGrid(s *game.Session) {
	s.IO.ANSICode(fmt.Sprintf("%d;%dH", s.UserR+2, s.UserC+2))
	s.IO.ANSICode("1;34m@") // blue @
}

func drawMonsterOnGrid(s *game.Session) {
	s.IO.ANSICode(fmt.Sprintf("%d;%dH", s.MonsR+2, s.MonsC+2))
	s.IO.ANSICode("1;31mM") // red M
}

func drawTerrainCell(s *game.Session, r, c int) {
	td := mechanics.GetTerrainDisplay(s.Terrain.Cells[r][c])
	s.IO.ANSICode(fmt.Sprintf("%d;%dH", r+2, c+2))
	s.IO.ANSICode(td.ANSIColor + td.Char)
}

func drawGridStats(s *game.Session) {
	c := s.Character
	col := model.GridCols + 4

	s.IO.ANSICode(fmt.Sprintf("2;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;36m%s", c.Name))

	s.IO.ANSICode(fmt.Sprintf("3;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;32mHP: %d/%d", c.HitPoints, c.MaxHP))

	s.IO.ANSICode(fmt.Sprintf("4;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;32mMv: %d", c.Movement))

	s.IO.ANSICode(fmt.Sprintf("5;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;32mPs: %d/%d", c.Psyche, c.MaxPsyche))

	s.IO.ANSICode(fmt.Sprintf("7;%dH", col))
	w1 := c.Weapons[0].Name
	if w1 == "" {
		w1 = "Hands"
	}
	s.IO.ANSICode(fmt.Sprintf("0;35mW1: %s", w1))

	s.IO.ANSICode(fmt.Sprintf("8;%dH", col))
	w2 := c.Weapons[1].Name
	if w2 == "" {
		w2 = "---"
	}
	s.IO.ANSICode(fmt.Sprintf("0;35mW2: %s", w2))

	s.IO.ANSICode(fmt.Sprintf("10;%dH", col))
	a1 := c.Armor[0].Name
	if a1 == "" {
		a1 = "None"
	}
	s.IO.ANSICode(fmt.Sprintf("0;35mAr: %s", a1))

	// Monster info
	if s.Monster != nil {
		s.IO.ANSICode(fmt.Sprintf("12;%dH", col))
		s.IO.ANSICode(fmt.Sprintf("0;31m%s", s.Monster.Name))
		s.IO.ANSICode(fmt.Sprintf("13;%dH", col))
		s.IO.ANSICode(fmt.Sprintf("0;31mHP: %d", s.Monster.HitPoints))
	}
}

func updateMovementDisplay(s *game.Session) {
	col := model.GridCols + 4
	s.IO.ANSICode(fmt.Sprintf("4;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;32mMv: %-4d", s.Character.Movement))

	// Update HP too
	s.IO.ANSICode(fmt.Sprintf("3;%dH", col))
	s.IO.ANSICode(fmt.Sprintf("0;32mHP: %d/%d  ", s.Character.HitPoints, s.Character.MaxHP))

	// Update monster HP
	if s.Monster != nil {
		s.IO.ANSICode(fmt.Sprintf("13;%dH", col))
		s.IO.ANSICode(fmt.Sprintf("0;31mHP: %-4d", s.Monster.HitPoints))
	}

	// Position cursor for input
	s.IO.ANSICode(fmt.Sprintf("%d;1H", model.GridRows+10))
}

func gridTextOut(s *game.Session, msg string) {
	s.TextOutln++
	if s.TextOutln > 5 {
		s.TextOutln = 1
	}
	row := model.GridRows + 3 + s.TextOutln
	// Pad to 50 chars
	padded := msg
	if len(padded) > 50 {
		padded = padded[:50]
	}
	padded = fmt.Sprintf("%-50s", padded)
	s.IO.ANSICode(fmt.Sprintf("%d;2H", row))
	s.IO.ANSICode("1;30;40m" + padded)
}

func moveGridPlayer(s *game.Session, dr, dc int) {
	newR := s.UserR + dr
	newC := s.UserC + dc

	// Bounds check
	if newR < 0 {
		newR = 0
	}
	if newR >= model.GridRows {
		newR = model.GridRows - 1
	}
	if newC < 0 {
		newC = 0
	}
	if newC >= model.GridCols {
		newC = model.GridCols - 1
	}

	cost := mechanics.NumMovePointsAt(newR, newC, s.Terrain)

	if cost >= 998 {
		// Water = instant death
		if s.Terrain.Cells[newR][newC] == model.TerrainWater {
			gridTextOut(s, "You fall into the water and drown!")
			s.Character.Alive = false
			s.Character.HitPoints = 0
			return
		}
		gridTextOut(s, "You can't move there!")
		return
	}

	if cost > s.Character.Movement {
		gridTextOut(s, "Not enough movement points!")
		s.Character.Movement--
		if s.Character.Movement < 0 {
			s.Character.Movement = 0
		}
		return
	}

	// Erase old position, draw terrain
	drawTerrainCell(s, s.UserR, s.UserC)

	s.UserR = newR
	s.UserC = newC
	s.Character.Movement -= cost

	// Draw player at new position
	drawPlayerOnGrid(s)
}

func doGridShoot(s *game.Session) {
	s.Character.Movement -= 3

	// Calculate distance
	dr := float64(s.MonsR - s.UserR)
	dc := float64(s.MonsC - s.UserC)
	dist := math.Sqrt(dr*dr + dc*dc)

	weapRange := s.Character.Weapons[1].Range
	if weapRange <= 0 {
		weapRange = s.Character.Weapons[0].Range
	}

	if weapRange <= 0 {
		gridTextOut(s, "You have no ranged weapon!")
		return
	}

	if int(dist) > weapRange {
		gridTextOut(s, "Your Shot Fell Short!")
		return
	}

	// Hit check: based on dexterity vs distance
	hitChance := s.Character.Dexterity*3 - int(dist)*2
	if randBetween(1, 100) > hitChance {
		misses := []string{
			"Your Shot Fell Short!",
			"Your Shot Bounces off a Boulder!",
			"Your Aim isn't very good, it hit a TREE!",
			"Aww.. Too Bad, Hit a Tree...",
			"Argh! Take some Lessons!",
		}
		gridTextOut(s, misses[randBetween(0, len(misses)-1)])
		return
	}

	// Damage based on ranged weapon
	strike := s.Character.Weapons[1].Strike
	if strike <= 0 {
		strike = s.Character.Weapons[0].Strike
	}
	damage := randBetween(strike/2, strike) + randBetween(0, s.Character.Strength/3)
	damage -= s.Monster.Defense / 2
	if damage < 1 {
		damage = 1
	}

	s.Monster.HitPoints -= damage
	gridTextOut(s, fmt.Sprintf("You hit %s for %d Damage!", s.Monster.Name, damage))

	// Flash at monster location
	s.IO.ANSICode(fmt.Sprintf("%d;%dH", s.MonsR+2, s.MonsC+2))
	s.IO.ANSICode("1;33m*")
	drawMonsterOnGrid(s)
}

func doGridMonsterShoot(s *game.Session) {
	if s.Monster.RangeOffense <= 0 {
		return
	}

	dr := float64(s.UserR - s.MonsR)
	dc := float64(s.UserC - s.MonsC)
	dist := math.Sqrt(dr*dr + dc*dc)

	if int(dist) > s.Monster.Range {
		return
	}

	// Hit check
	if randBetween(1, 100) > 60 {
		gridTextOut(s, fmt.Sprintf("The %s shoots and misses!", s.Monster.Name))
		return
	}

	damage := randBetween(s.Monster.RangeOffense/2, s.Monster.RangeOffense) - s.Character.Armor[0].Defense/2
	if damage < 1 {
		damage = 1
	}

	s.Character.HitPoints -= damage
	gridTextOut(s, fmt.Sprintf("The %s hits you for %d ranged damage!", s.Monster.Name, damage))
}
