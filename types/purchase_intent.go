package types

import "time"

type PurchaseIntentStatus string

const (
	PurchaseIntentPending   PurchaseIntentStatus = "pending"
	PurchaseIntentAccepted  PurchaseIntentStatus = "accepted"
	PurchaseIntentRejected  PurchaseIntentStatus = "rejected"
	PurchaseIntentCanceled  PurchaseIntentStatus = "canceled"
	PurchaseIntentCompleted PurchaseIntentStatus = "completed"
)

type PurchaseIntent struct {
	ID          string               `json:"id"`
	UserID      string               `json:"user_id"`
	Product     Product              `json:"product"`
	OfferPrice  int64                `json:"offer_price"`
	Status      PurchaseIntentStatus `json:"status"`
	PickupNotes string               `json:"pickup_notes"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
