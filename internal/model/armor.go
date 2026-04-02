package model

// ArmorClass represents what body slot armor occupies.
type ArmorClass int

const (
	ArmorBody     ArmorClass = 0 // body armor
	ArmorShield   ArmorClass = 1 // shield
	ArmorFullBody ArmorClass = 2 // full body armor
)

// Armor represents armor that can be purchased or worn.
// Original: ArmorRec in SlyHeaders.h
type Armor struct {
	Name      string     `json:"name"`
	Defense   int        `json:"defense"`   // protection rating
	Class     ArmorClass `json:"class"`     // 0=body, 1=shield, 2=full body
	Cost      int        `json:"cost"`      // purchase price
	Ability   int        `json:"ability"`   // special ability index
	Quality   int        `json:"quality"`   // armor quality rating
	ActionStr string     `json:"actionStr"` // protective verb (blocks, parries)
}

// IsEmpty returns true if this armor slot is unequipped.
func (a Armor) IsEmpty() bool {
	return a.Name == ""
}
