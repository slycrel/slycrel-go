package model

// Coord represents a position on the combat grid.
type Coord struct {
	X int `json:"x"` // column (0-47)
	Y int `json:"y"` // row (0-11)
}

// CellRec represents pathfinding data for one cell on the combat grid.
type CellRec struct {
	Weight    int  `json:"weight"`
	Direction int  `json:"direction"`
	Chosen    bool `json:"chosen"`
}

// TerrainMap represents a 12x48 combat terrain grid.
type TerrainMap struct {
	Name  string              `json:"name"`
	Cells [12][48]TerrainCell `json:"cells"`
}

// Combat grid dimensions (from original).
const (
	GridRows = 12
	GridCols = 48
)

// Game balance constants from SlyHeaders.h.
const (
	BaseExp           = 50 // initial level 1 XP requirement
	MaxExploration    = 15 // daily exploration limit
	MaxSpars          = 4  // daily spar limit
	StartingCoinsHand = 15
	StartingCoinsBank = 5
)
