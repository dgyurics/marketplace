package types

// MailjetConfig holds the configuration for the Mailjet email sender
type MailjetConfig struct {
	Enabled   bool   // whether to enable email sending
	APIKey    string // Mailjet API key
	APISecret string // Mailjet API secret
	FromEmail string // email address to send from (must be verified in Mailjet)
	FromName  string // name to send from	(eg. "Marketplace")
}
