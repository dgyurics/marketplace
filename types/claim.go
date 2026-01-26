package types

type Claim struct {
	ID          string  `json:"id"`
	UserID      string  `json:"-"`
	PickupNotes string  `json:"pickup_notes"`
	Product     Product `json:"product"`
}
