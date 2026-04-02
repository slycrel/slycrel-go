package model

// Monster represents a creature that can be fought.
// Original: MonsterRec in SlyHeaders.h
type Monster struct {
	Name             string      `json:"name"`
	Level            int         `json:"level"`
	MonType          MonsterType `json:"monType"`
	MonsterNum       int         `json:"monsterNum"`
	HitPoints        int         `json:"hitPoints"`
	Offense          int         `json:"offense"`
	RangeOffense     int         `json:"rangeOffense"`
	Range            int         `json:"range"`
	Defense          int         `json:"defense"`
	Movement         int         `json:"movement"`
	Disposition      int         `json:"disposition"`
	Experience       int         `json:"experience"`       // XP reward
	Coins            int         `json:"coins"`            // gold reward
	AttackStr1       string      `json:"attackStr1"`       // primary attack description
	AttackStr2       string      `json:"attackStr2"`       // secondary attack description
	AttackActionStr  string      `json:"attackActionStr"`  // attack verb (wave, send, point)
	DefenseStr       string      `json:"defenseStr"`       // defense description
	DefenseActionStr string      `json:"defenseActionStr"` // defense verb
}
