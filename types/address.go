package types

import "time"

type Address struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Country    string    `json:"country"` // FIXME this is country code
	Name       *string   `json:"name,omitempty"`
	Line1      string    `json:"line1"`
	Line2      *string   `json:"line2,omitempty"`
	City       string    `json:"city"`            // city, district, suburb, town, village
	State      *string   `json:"state,omitempty"` // state, county, province, region
	PostalCode string    `json:"postal_code"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ShippingZone struct {
	ID         string  `json:"id"`
	Country    string  `json:"country"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
}

type ExcludedShippingZone struct {
	ID         string `json:"id"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
}
