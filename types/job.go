package types

type Job string

const (
	StaleOrders             Job = "stale_orders"
	StateAddresses          Job = "stale_addresses"
	StalePasswordResetCodes Job = "stale_reset_codes"
	StaleRefreshTokens      Job = "stale_refresh_tokens"
)
