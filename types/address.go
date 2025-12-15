package types

// TODO move address to this file

// id BIGINT PRIMARY KEY,
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
