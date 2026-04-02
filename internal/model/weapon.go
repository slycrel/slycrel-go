package model

// Weapon represents a weapon that can be purchased or wielded.
// Original: WeaponRec in SlyHeaders.h
type Weapon struct {
	Name      string `json:"name"`
	Strike    int    `json:"strike"`    // damage rating
	Range     int    `json:"range"`     // weapon range (0 = melee)
	Cost      int    `json:"cost"`      // purchase price
	Ability   int    `json:"ability"`   // special ability index
	Quality   int    `json:"quality"`   // weapon quality rating
	ActionStr string `json:"actionStr"` // verb (slash, pierce, etc.)
}

// IsEmpty returns true if this weapon slot is unequipped.
func (w Weapon) IsEmpty() bool {
	return w.Name == ""
}
