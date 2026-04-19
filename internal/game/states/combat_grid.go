package states

import (
	"fmt"
	"math"

	"github.com/slycrel/slycrel/internal/game"
	slyio "github.com/slycrel/slycrel/internal/io"
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
	renderGridScene(s)
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

	// Refresh display (HP, movement, positions)
	renderGridScene(s)

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
		renderGridScene(s)
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

		// Redraw scene at monster's new position
		renderGridScene(s)
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
//
// All rendering flows through renderGridScene: build a Scene from current
// session state, hand it to the IOProvider. LocalTerminal renders ANSI as
// before; WSSession ships structured data to tilemap clients (Godot).
// Incremental "draw a single cell" updates are gone — we always re-render
// the full scene, since 12×48 is tiny and one less code path to maintain.

func renderGridScene(s *game.Session) {
	if s.Terrain == nil {
		return
	}
	c := s.Character
	hud := slyio.SceneHUD{
		CharName:  c.Name,
		HP:        c.HitPoints,
		MaxHP:     c.MaxHP,
		Movement:  c.Movement,
		Psyche:    c.Psyche,
		MaxPsyche: c.MaxPsyche,
		Weapon1:   c.Weapons[0].Name,
		Weapon2:   c.Weapons[1].Name,
		Armor:     c.Armor[0].Name,
		Log:       activeGridLog(s),
	}
	if s.Monster != nil {
		hud.MonsterName = s.Monster.Name
		hud.MonsterHP = s.Monster.HitPoints
	}
	scene := slyio.NewGridCombatScene(s.Terrain, s.UserR, s.UserC, s.MonsR, s.MonsC, hud)
	s.IO.RenderScene(scene)
}

// activeGridLog returns the status log in most-recent-last order by walking
// the ring buffer from the oldest slot. Empty slots are dropped so early
// turns don't show blank lines.
func activeGridLog(s *game.Session) []string {
	out := make([]string, 0, len(s.GridStatusMsgs))
	start := s.TextOutln % len(s.GridStatusMsgs)
	for i := 0; i < len(s.GridStatusMsgs); i++ {
		idx := (start + i) % len(s.GridStatusMsgs)
		if msg := s.GridStatusMsgs[idx]; msg != "" {
			out = append(out, msg)
		}
	}
	return out
}

func gridTextOut(s *game.Session, msg string) {
	s.GridStatusMsgs[s.TextOutln%len(s.GridStatusMsgs)] = msg
	s.TextOutln++
	renderGridScene(s)
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

	s.UserR = newR
	s.UserC = newC
	s.Character.Movement -= cost

	renderGridScene(s)
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
	// Note: dropped the single-frame yellow `*` flash at the monster's cell —
	// gridTextOut already triggers a full re-render via renderGridScene.
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
