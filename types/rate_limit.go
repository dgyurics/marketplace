package types

import (
	"time"
)

type RateLimit struct {
	IPAddress string // net.IP is available, but makes storing/retrieving from DB a pain
	Path      string
	HitCount  int
	Limit     int
	ExpiresAt time.Time
}
