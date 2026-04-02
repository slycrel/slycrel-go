package model

// RoomRec represents a single room in the inn.
// Original: RoomRec in SlyHeaders.h
type RoomRec struct {
	Who      string `json:"who"`      // occupant name
	DaysLeft int    `json:"daysLeft"` // rental days remaining
	Lock     int    `json:"lock"`     // lock quality
}

// InnRec represents the entire inn state.
// Original: InnRec in SlyHeaders.h
type InnRec struct {
	Rooms   [10]RoomRec `json:"rooms"`
	Owner   string      `json:"owner"`
	CurRate int         `json:"curRate"` // daily room rate
	Safe    int64       `json:"safe"`    // gold in the innkeeper's safe
	Open    bool        `json:"open"`    // is the inn open for business
}
