package mechanics

import "github.com/slycrel/slycrel/internal/model"

// TerrainDisplay holds the ANSI color and character for a terrain type.
type TerrainDisplay struct {
	ANSIColor string
	Char      string
}

// GetTerrainDisplay returns the display info for a terrain cell.
// Original: GetTerrChar() in Sly_CombatUnit.cp
var terrainDisplays = [10]TerrainDisplay{
	{ANSIColor: "0;37;40m", Char: " "},  // 0 - Empty/grass
	{ANSIColor: "0;32;40m", Char: "^"},  // 1 - Forest (° in original, using ^ for terminal compat)
	{ANSIColor: "0;33;40m", Char: "."},  // 2 - Sand/road (° in original)
	{ANSIColor: "0;34;44m", Char: "~"},  // 3 - Water (÷ in original)
	{ANSIColor: "0;33;43m", Char: "="},  // 4 - Bridge (ð in original)
	{ANSIColor: "1;30;40m", Char: "O"},  // 5 - Boulder (í in original)
	{ANSIColor: "1;32;40m", Char: "#"},  // 6 - Deep forest (\006 in original)
	{ANSIColor: "0;32;40m", Char: "+"},  // 7 - Light forest
	{ANSIColor: "1;33;42m", Char: "%"},  // 8 - Swamp (± in original)
	{ANSIColor: "0;32;42m", Char: "&"},  // 9 - Deep swamp (² in original)
}

// GetTerrainDisplay returns display info for the terrain type.
func GetTerrainDisplay(cell model.TerrainCell) TerrainDisplay {
	if int(cell) < len(terrainDisplays) {
		return terrainDisplays[cell]
	}
	return TerrainDisplay{ANSIColor: "1;31;40m", Char: "?"}
}

// NumMovePoints returns the movement cost to enter a cell.
// Original: NumMovePoints() in Sly_CombatUnit.cp
func NumMovePoints(cell model.TerrainCell) int {
	switch cell {
	case model.TerrainEmpty:
		return 1 // grass
	case model.TerrainPlainGr:
		return 2 // forest
	case model.TerrainPlainBr:
		return 1 // sand/road
	case model.TerrainWater:
		return 998 // impassable (instant death)
	case model.TerrainBridge:
		return 1
	case model.TerrainBoulder:
		return 998 // impassable
	case model.TerrainForest:
		return 3 // thick forest
	case model.TerrainDeepForest:
		return 4 // deep forest
	case model.TerrainSwamp:
		return 3
	case model.TerrainDeepSwamp:
		return 5
	default:
		return 998
	}
}

// NumMovePointsAt returns the movement cost for a grid position.
func NumMovePointsAt(r, c int, terrain *model.TerrainMap) int {
	if r < 0 || r >= model.GridRows || c < 0 || c >= model.GridCols {
		return 998
	}
	return NumMovePoints(terrain.Cells[r][c])
}

// CanMoveAnywhere checks if the player can move in any direction.
// Original: CanMove() in Sly_CombatUnit.cp
func CanMoveAnywhere(r, c, movement int, terrain *model.TerrainMap) bool {
	if r > 0 && movement >= NumMovePointsAt(r-1, c, terrain) {
		return true
	}
	if r < model.GridRows-1 && movement >= NumMovePointsAt(r+1, c, terrain) {
		return true
	}
	if c > 0 && movement >= NumMovePointsAt(r, c-1, terrain) {
		return true
	}
	if c < model.GridCols-1 && movement >= NumMovePointsAt(r, c+1, terrain) {
		return true
	}
	return false
}

// FindRoute calculates a path from (startR, startC) to (destR, destC).
// Returns a string of U/D/L/R directions and whether a path was found.
// Original: FindRoute() - Dijkstra-style pathfinding in Sly_CombatUnit.cp
func FindRoute(startR, startC, destR, destC int, terrain *model.TerrainMap) (string, bool) {
	type cell struct {
		weight    int
		direction byte // 'U', 'D', 'L', 'R'
		visited   bool
	}

	var grid [model.GridRows][model.GridCols]cell
	for r := 0; r < model.GridRows; r++ {
		for c := 0; c < model.GridCols; c++ {
			grid[r][c].weight = 32700
		}
	}

	grid[startR][startC].weight = NumMovePointsAt(startR, startC, terrain)

	// Priority queue as simple list (good enough for 12x48)
	type queueItem struct {
		r, c int
	}
	queue := []queueItem{{startR, startC}}

	maxIter := 1050
	for iter := 0; iter < maxIter && len(queue) > 0; iter++ {
		// Find lowest weight in queue
		bestIdx := 0
		for i, item := range queue {
			if grid[item.r][item.c].weight < grid[queue[bestIdx].r][queue[bestIdx].c].weight {
				bestIdx = i
			}
		}
		cur := queue[bestIdx]
		queue = append(queue[:bestIdx], queue[bestIdx+1:]...)
		grid[cur.r][cur.c].visited = true

		if cur.r == destR && cur.c == destC {
			break
		}

		// Check 4 neighbors
		neighbors := [4]struct {
			r, c int
			dir  byte
		}{
			{cur.r - 1, cur.c, 'U'},
			{cur.r + 1, cur.c, 'D'},
			{cur.r, cur.c - 1, 'L'},
			{cur.r, cur.c + 1, 'R'},
		}

		for _, n := range neighbors {
			if n.r < 0 || n.r >= model.GridRows || n.c < 0 || n.c >= model.GridCols {
				continue
			}
			if grid[n.r][n.c].visited {
				continue
			}
			cost := grid[cur.r][cur.c].weight + NumMovePointsAt(n.r, n.c, terrain)
			if cost < grid[n.r][n.c].weight {
				grid[n.r][n.c].weight = cost
				grid[n.r][n.c].direction = n.dir
				queue = append(queue, queueItem{n.r, n.c})
			}
		}
	}

	// Backtrack from destination to start
	if !grid[destR][destC].visited && grid[destR][destC].weight == 32700 {
		return "", false
	}

	var route []byte
	r, c := destR, destC
	for r != startR || c != startC {
		dir := grid[r][c].direction
		route = append([]byte{dir}, route...)
		switch dir {
		case 'U':
			r++
		case 'D':
			r--
		case 'L':
			c++
		case 'R':
			c--
		}
		if len(route) > 100 {
			break // safety
		}
	}

	return string(route), true
}
