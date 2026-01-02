package types

type Job string

const (
	StaleOrders              Job = "stale_orders"
	StateAddresses           Job = "stale_addresses"
	StaleCartItems           Job = "stale_cart_items"
	ExpiredRateLimits        Job = "expired_rate_limits"
	ExpiredRegistrationCodes Job = "expired_registration_codes"
	ExpiredRefreshTokens     Job = "expired_refresh_tokens"
	ExpiredPasswordResets    Job = "expired_password_resets"
)
