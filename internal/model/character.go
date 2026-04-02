package model

// Character holds all player data. Faithfully recreates CharacterRec from SlyHeaders.h.
type Character struct {
	FileIndex          int64     `json:"fileIndex"`
	Name               string    `json:"name"`               // in-game character name
	BBSName            string    `json:"bbsName"`             // login username
	ScoreVal           int64     `json:"scoreVal"`            // calculated appraisal value
	LastOn             int64     `json:"lastOn"`              // last login date (YYYYMMDD)
	Alive              bool      `json:"alive"`
	Location           WhereType `json:"location"`
	Gender             bool      `json:"gender"`              // true=male, false=female
	CharClass          CharClass `json:"charClass"`
	TotalExperience    int64     `json:"totalExperience"`
	SpendingExperience int64     `json:"spendingExperience"`
	ThiefLvl           int       `json:"thiefLvl"`
	MageLvl            int       `json:"mageLvl"`
	FighterLvl         int       `json:"fighterLvl"`
	HitPoints          int       `json:"hitPoints"`
	MaxHP              int       `json:"maxHP"`
	CoinsHand          int64     `json:"coinsHand"`
	CoinsBank          int64     `json:"coinsBank"`
	Speed              int       `json:"speed"`
	Strength           int       `json:"strength"`
	Dexterity          int       `json:"dexterity"`
	Psyche             int       `json:"psyche"`
	MaxPsyche          int       `json:"maxPsyche"`
	Fame               int       `json:"fame"`
	Honor              int       `json:"honor"`
	Faith              int       `json:"faith"`
	FavColor           Color     `json:"favColor"`
	Flirt1             int       `json:"flirt1"`
	Flirt2             int       `json:"flirt2"`
	Ammo               int       `json:"ammo"`
	Movement           int       `json:"movement"`
	Exploration        int       `json:"exploration"`        // daily limit: 15
	Spars              int       `json:"spars"`              // daily limit: 4
	TotalFights        int64     `json:"totalFights"`
	FightsWon          int64     `json:"fightsWon"`
	Weapons            [3]Weapon `json:"weapons"`            // [0]=melee, [1]=ranged, [2]=extra
	Armor              [2]Armor  `json:"armor"`              // [0]=body, [1]=shield
}

// GenderString returns "Male" or "Female".
func (c *Character) GenderString() string {
	if c.Gender {
		return "Male"
	}
	return "Female"
}

// Level returns the character's primary class level.
func (c *Character) Level() int {
	switch c.CharClass {
	case ClassFighter:
		return c.FighterLvl
	case ClassThief:
		return c.ThiefLvl
	case ClassMage:
		return c.MageLvl
	default:
		return c.FighterLvl
	}
}

// HighScore represents a high score entry.
// Original: HighScoreType in SlyHeaders.h
type HighScore struct {
	Name     string `json:"name"`
	Value    int64  `json:"value"`
	LastDate int64  `json:"lastDate"`
}
