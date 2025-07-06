package types

// TODO init endpoint which returns locale data for country and currency
// specified during setup

// E.g. GET /api/locale
// type LocaleConfig struct {
//     Country         string `json:"country"`          // "US", "DK", "JP"
//     CountryName     string `json:"country_name"`     // "United States", "Denmark", "Japan"
//     DateFormat      string `json:"date_format"`      // "MM/DD/YYYY", "DD/MM/YYYY"
//     DecimalSep      string `json:"decimal_separator"`// ".", ","
//     ThousandsSep    string `json:"thousands_separator"` // ",", "."
//     Currency        string `json:"currency"`         // "USD", "DKK", "JPY"
//     CurrencyName    string `json:"currency_name"`    // "US Dollar", "Danish Krone", "Japanese Yen"
//     CurrencySymbol  string `json:"currency_symbol"`  // "$", "kr", "¥"
//     Multiplier      int    `json:"multiplier"`       // 100, 100, 1
//     SmallestUnit    string `json:"smallest_unit"`    // "cents", "øre", "yen"
//     PostalPattern   string `json:"postal_pattern"`   // Regex for postal code validation
// }

// User enters: 19.99
// Backend calculation: 19.99 * 100 = 1999
// Database stores: 1999 (integer)
// Display: "$19.99"
// From 1999 we are able to display $19.99
// using SymbolFirst = true
// Multiplier = 100
// DecimalSep = "."
// ThousandsSep = ","

// For DKK:
// User enters: 19.99
// Backend calculation: 19.99 * 100 = 1999
// Database stores: 1999 (integer)
// Display: "19,99 kr"
// From 1999 we are able to display 19,99 kr
// using SymbolFirst = false
// Multiplier = 100
// DecimalSep = ","
// ThousandsSep = "."

// For JPY:
// User enters: 1999
// Backend calculation: 1999 * 1 = 1999
// Database stores: 1999 (integer)
// Display: "¥1,999"
// From 1999 we are able to display ¥1,999
// using SymbolFirst = true
// Multiplier = 1
// DecimalSep = "."
// ThousandsSep = ","

// TODO use in conjunction with SupportedCountries map
type Country struct {
	CountryCode   string `json:"country_code"`   // ISO 3166-1 alpha-2 "US", "DK", "JP"
	Name          string `json:"name"`           // "United States", "Denmark", "Japan"
	DateFormat    string `json:"date_format"`    // "MM/DD/YYYY", "DD/MM/YYYY", etc.
	PostalPattern string `json:"postal_pattern"` // Regex for postal code validation
	DecimalSep    string `json:"decimal_sep"`    // ".", ","
	ThousandsSep  string `json:"thousands_sep"`  // ",", ".", " "
}

// TODO use in conjunction with CurrencyConfig map
type Currency struct {
	Code         string `json:"code"`          // ISO 4217 "USD", "DKK", "JPY"
	Name         string `json:"name"`          // "US Dollar", "Danish Krone", "Japanese Yen"
	Symbol       string `json:"symbol"`        // "$", "kr", "¥"
	Multiplier   int    `json:"multiplier"`    // 100, 100, 1
	SmallestUnit string `json:"smallest_unit"` // "cents", "øre", "yen"
	SymbolFirst  bool   `json:"symbol_first"`  // true for "$19.99", false for "19.99 kr"
}
