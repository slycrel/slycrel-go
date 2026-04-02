package model

// GCombatPrefs represents an active gladiator fight.
// Original: GCombatPrefs in SlyHeaders.h
type GCombatPrefs struct {
	Challenger string `json:"challenger"`
	Opponent   string `json:"opponent"`
	Odds       [2]int `json:"odds"` // [0]=challenger odds, [1]=opponent odds
}

// GCombatBet represents a bet on a gladiator fight.
// Original: GCombatList in SlyHeaders.h
type GCombatBet struct {
	FightNum       int    `json:"fightNum"`       // which fight the bet is on
	FileName       string `json:"fileName"`       // bettor's character file name
	Bet            int    `json:"bet"`            // amount wagered
	Challenger     string `json:"challenger"`     // challenger name (for display)
	Opponent       string `json:"opponent"`       // opponent name (for display)
	ForChallenger  bool   `json:"forChallenger"`  // true=bet on challenger, false=bet on opponent
}
