package model

// TownConfig holds global game configuration.
// Original: GeneralRec in SlyHeaders.h
type TownConfig struct {
	TownName          string   `json:"townName"`
	NumberPlayers     int      `json:"numberPlayers"`
	GoldPool          int64    `json:"goldPool"`
	ManaPool          int64    `json:"manaPool"`
	RealTime          bool     `json:"realTime"`
	Population        int64    `json:"population"`
	TownSize          TownType `json:"townSize"`
	TaxRate           float64  `json:"taxRate"`           // percentage
	PlayerPower       bool     `json:"playerPower"`
	NPCNum            int      `json:"npcNum"`            // ceiling of # players * 2
	Difficulty        float64  `json:"difficulty"`         // 0.5 to 3.0 multiplier, default 1.0
	LevelUpDifficulty int      `json:"levelUpDifficulty"` // 1-15, default 4
}

// DefaultTownConfig returns a sensible default town configuration.
func DefaultTownConfig() TownConfig {
	return TownConfig{
		TownName:          "Slycrel",
		NumberPlayers:     0,
		GoldPool:          10000,
		ManaPool:          5000,
		RealTime:          false,
		Population:        500,
		TownSize:          SmallTown,
		TaxRate:           0.05,
		PlayerPower:       false,
		NPCNum:            0,
		Difficulty:        1.0,
		LevelUpDifficulty: 4,
	}
}
