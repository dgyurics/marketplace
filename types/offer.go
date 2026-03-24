package types

import "time"

type OfferStatus string

const (
	OfferPending   OfferStatus = "pending"
	OfferAccepted  OfferStatus = "accepted"
	OfferRejected  OfferStatus = "rejected"
	OfferCanceled  OfferStatus = "canceled"
	OfferCompleted OfferStatus = "completed"
)

type Offer struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	Product     Product     `json:"product"`
	Amount      int64       `json:"amount"`
	Status      OfferStatus `json:"status"`
	PickupNotes string      `json:"pickup_notes"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
