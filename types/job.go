package types

type Job string

const (
	StaleOrders              Job = "stale_orders"
	StaleRequestedOrders     Job = "stale_requested_orders"
	StaleAddresses           Job = "stale_addresses"
	StaleGuestUsers          Job = "stale_guest_users"
	StaleCartItems           Job = "stale_cart_items"
	ExpiredRateLimits        Job = "expired_rate_limits"
	ExpiredRegistrationCodes Job = "expired_registration_codes"
	ExpiredRefreshTokens     Job = "expired_refresh_tokens"
	ExpiredPasswordResets    Job = "expired_password_resets"
)
