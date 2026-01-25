package types

type Claim struct {
	ID          string  `json:"id"`
	UserID      string  `json:"-"`
	PickupNotes string  `json:"pickup-notes"`
	Product     Product `json:"product"`
}
