package utilities

import (
	"errors"
	"regexp"
)

type Locale struct {
	Country           string
	PostalCodeLabel   string
	PostalCodePattern string
	StateLabel        string
	StateRequired     bool
	StateCodes        map[string]string
}

var LocaleData = map[string]Locale{
	"US": {
		Country:           "United States",
		PostalCodeLabel:   "ZIP Code",
		PostalCodePattern: PostalCodePatterns["US"],
		StateLabel:        "State",
		StateRequired:     true,
		StateCodes:        StateNames["US"],
	},
	"CA": {
		Country:           "Canada",
		PostalCodeLabel:   "Postal Code",
		PostalCodePattern: PostalCodePatterns["CA"],
		StateLabel:        "Province",
		StateRequired:     true,
		StateCodes:        StateNames["CA"],
	},
	"GB": {
		Country:           "United Kingdom",
		PostalCodeLabel:   "Postcode",
		PostalCodePattern: PostalCodePatterns["GB"],
		StateLabel:        "County",
		StateRequired:     false,
		StateCodes:        nil,
	},
	"DE": {
		Country:           "Germany",
		PostalCodeLabel:   "Postal Code",
		PostalCodePattern: PostalCodePatterns["DE"],
		StateLabel:        "State",
		StateRequired:     false,
		StateCodes:        nil,
	},
	// Additional countries can be added here
}

// Supported ISO 3166-1 alpha-2 countries
var SupportedCountries = map[string]bool{
	"AE": true,
	"AR": true,
	"AT": true,
	"AU": true,
	"BE": true,
	"BR": true,
	"CA": true,
	"CH": true,
	"CI": true,
	"CL": true,
	"CO": true,
	"CY": true,
	"CZ": true,
	"DE": true,
	"DK": true,
	"EE": true,
	"EG": true,
	"ES": true,
	"FI": true,
	"FR": true,
	"GB": true,
	"GH": true,
	"GI": true,
	"GR": true,
	"HK": true,
	"HR": true,
	"HU": true,
	"ID": true,
	"IE": true,
	"IL": true,
	"IN": true,
	"IS": true,
	"IT": true,
	"JP": true,
	"KE": true,
	"KR": true,
	"LI": true,
	"LK": true,
	"LT": true,
	"LU": true,
	"LV": true,
	"MA": true,
	"MT": true,
	"MX": true,
	"MY": true,
	"NG": true,
	"NL": true,
	"NO": true,
	"NZ": true,
	"PA": true,
	"PE": true,
	"PH": true,
	"PL": true,
	"PT": true,
	"RO": true,
	"SA": true,
	"SE": true,
	"SG": true,
	"SI": true,
	"SK": true,
	"TH": true,
	"TW": true,
	"US": true,
	"UY": true,
	"VN": true,
	"ZA": true,
}

// Supported ISO 4217 currencies
var SupportedCurrencies = map[string]bool{
	"AED": true,
	"AFN": true,
	"ALL": true,
	"AMD": true,
	"ANG": true,
	"AOA": true,
	"ARS": true,
	"AUD": true,
	"AWG": true,
	"AZN": true,
	"BAM": true,
	"BBD": true,
	"BDT": true,
	"BIF": true,
	"BMD": true,
	"BND": true,
	"BOB": true,
	"BRL": true,
	"BSD": true,
	"BWP": true,
	"BZD": true,
	"CAD": true,
	"CDF": true,
	"CHF": true,
	"CLP": true,
	"CNY": true,
	"COP": true,
	"CRC": true,
	"CVE": true,
	"CZK": true,
	"DJF": true,
	"DKK": true,
	"DOP": true,
	"DZD": true,
	"EGP": true,
	"ETB": true,
	"EUR": true,
	"FJD": true,
	"FKP": true,
	"GBP": true,
	"GEL": true,
	"GIP": true,
	"GMD": true,
	"GNF": true,
	"GTQ": true,
	"GYD": true,
	"HKD": true,
	"HNL": true,
	"HUF": true,
	"IDR": true,
	"ILS": true,
	"INR": true,
	"ISK": true,
	"JMD": true,
	"JPY": true,
	"KES": true,
	"KGS": true,
	"KHR": true,
	"KMF": true,
	"KRW": true,
	"KYD": true,
	"KZT": true,
	"LAK": true,
	"LBP": true,
	"LKR": true,
	"LRD": true,
	"LSL": true,
	"MAD": true,
	"MDL": true,
	"MGA": true,
	"MKD": true,
	"MMK": true,
	"MNT": true,
	"MOP": true,
	"MRU": true,
	"MUR": true,
	"MVR": true,
	"MWK": true,
	"MXN": true,
	"MYR": true,
	"MZN": true,
	"NAD": true,
	"NGN": true,
	"NIO": true,
	"NOK": true,
	"NPR": true,
	"NZD": true,
	"PAB": true,
	"PEN": true,
	"PGK": true,
	"PHP": true,
	"PKR": true,
	"PLN": true,
	"PYG": true,
	"QAR": true,
	"RON": true,
	"RWF": true,
	"SAR": true,
	"SBD": true,
	"SCR": true,
	"SEK": true,
	"SGD": true,
	"SHP": true,
	"SLL": true,
	"SOS": true,
	"SRD": true,
	"STN": true,
	"SZL": true,
	"THB": true,
	"TJS": true,
	"TOP": true,
	"TRY": true,
	"TTD": true,
	"TWD": true,
	"TZS": true,
	"UAH": true,
	"UGX": true,
	"USD": true,
	"UYU": true,
	"UZS": true,
	"VES": true,
	"VND": true,
	"VUV": true,
	"WST": true,
	"XAF": true,
	"XCD": true,
	"XOF": true,
	"XPF": true,
	"YER": true,
	"ZAR": true,
	"ZMW": true,
}

var PostalCodePatterns = map[string]string{
	"AE": `.*`,                                    // UAE: Not mandatory
	"AR": `^([A-Z]\d{4}[A-Z]{3})|(\d{4})$`,        // Argentina C1425ABC or 1425
	"AT": `^\d{4}$`,                               // Austria 1234
	"AU": `^\d{4}$`,                               // Australia 4000
	"BE": `^\d{4}$`,                               // Belgium 1234
	"BR": `^\d{5}-\d{3}$`,                         // Brazil 12345-678
	"CA": `^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$`, // Canada A1A 1A1
	"CH": `^\d{4}$`,                               // Switzerland 1234
	"CI": `.*`,                                    // Côte d’Ivoire: no standard
	"CL": `^\d{7}$`,                               // Chile 8320000
	"CO": `^\d{6}$`,                               // Colombia 110111
	"CY": `^\d{4}$`,                               // Cyprus 1100
	"CZ": `^\d{3} ?\d{2}$`,                        // Czech Republic 110 00
	"DE": `^\d{5}$`,                               // Germany 12345
	"DK": `^\d{4}$`,                               // Denmark 1234
	"EE": `^\d{5}$`,                               // Estonia 12345
	"EG": `^\d{5}$`,                               // Egypt 12345
	"ES": `^\d{5}$`,                               // Spain 12345
	"FI": `^\d{5}$`,                               // Finland 12345
	"FR": `^\d{5}$`,                               // France 12345
	"GB": `^[A-Z]{1,2}\d[A-Z\d]? ?\d[A-Z]{2}$`,    // United Kingdom SW1A 1AA
	"GH": `.*`,                                    // Ghana: optional (GhanaPost GPS flexible)
	"GI": `^GX11 1AA$`,                            // Gibraltar GX11 1AA
	"GR": `^\d{3} ?\d{2}$`,                        // Greece 123 45
	"HK": `.*`,                                    // Hong Kong: no postal codes
	"HR": `^\d{5}$`,                               // Croatia 12345
	"HU": `^\d{4}$`,                               // Hungary 1011
	"ID": `^\d{5}$`,                               // Indonesia 12345
	"IE": `^[A-Za-z0-9]{3} ?[A-Za-z0-9]{4}$`,      // Ireland D02 X285
	"IL": `^\d{7}$`,                               // Israel 6100001
	"IN": `^\d{6}$`,                               // India 110001
	"IS": `^\d{3}$`,                               // Iceland 123
	"IT": `^\d{5}$`,                               // Italy 12345
	"JP": `^\d{3}-\d{4}$`,                         // Japan 123-4567
	"KE": `^\d{5}$`,                               // Kenya 00100
	"KR": `^\d{5}$`,                               // South Korea 12345
	"LI": `^\d{4}$`,                               // Liechtenstein 9490
	"LK": `^\d{5}$`,                               // Sri Lanka (placeholder)
	"LT": `^\d{5}$`,                               // Lithuania 12345
	"LU": `^\d{4}$`,                               // Luxembourg 1234
	"LV": `^\d{4}$`,                               // Latvia 1234
	"MA": `^\d{5}$`,                               // Morocco 10000
	"MT": `^[A-Z]{3} ?\d{4}$`,                     // Malta MLA 1001
	"MX": `^\d{5}$`,                               // Mexico 12345
	"MY": `^\d{5}$`,                               // Malaysia 43000
	"NG": `^\d{6}$`,                               // Nigeria 100001
	"NL": `^\d{4} ?[A-Z]{2}$`,                     // Netherlands 1234 AB
	"NO": `^\d{4}$`,                               // Norway 1234
	"NZ": `^\d{4}$`,                               // New Zealand 6011
	"PA": `.*`,                                    // Panama: no standard
	"PE": `^\d{5}$`,                               // Peru 15001
	"PH": `^\d{4}$`,                               // Philippines 1000
	"PL": `^\d{2}-\d{3}$`,                         // Poland 12-345
	"PT": `^\d{4}-\d{3}$`,                         // Portugal 1234-567
	"RO": `^\d{6}$`,                               // Romania 123456
	"SA": `^\d{5}$`,                               // Saudi Arabia (example: 11564)
	"SE": `^\d{3} ?\d{2}$`,                        // Sweden 123 45
	"SG": `^\d{6}$`,                               // Singapore 560123
	"SI": `^\d{4}$`,                               // Slovenia 1234
	"SK": `^\d{3} ?\d{2}$`,                        // Slovakia 123 45
	"TH": `^\d{5}$`,                               // Thailand 10110
	"TW": `^\d{3}(-\d{2})?$`,                      // Taiwan 123 or 123-45
	"US": `^\d{5}(-\d{4})?$`,                      // United States 12345 or 12345-6789
	"UY": `^\d{5}$`,                               // Uruguay 11300
	"VN": `^\d{6}$`,                               // Vietnam 700000
	"ZA": `^\d{4}$`,                               // South Africa 2000
}

func ValidatePostalCode(country, postalCode string) error {
	regex, ok := PostalCodePatterns[country]
	if !ok {
		return nil
	}
	if regexp.MustCompile(regex).MatchString(postalCode) {
		return nil
	}
	return errors.New("invalid postal code format")
}

func ValidateState(country, state string) error {
	states, ok := StateNames[country]
	if !ok {
		return nil
	}
	state, ok = states[state]
	if !ok {
		return errors.New("invalid state")
	}
	return nil
}

var StateNames = map[string]map[string]string{
	"US": {
		"AL": "Alabama", "AK": "Alaska", "AZ": "Arizona", "AR": "Arkansas",
		"CA": "California", "CO": "Colorado", "CT": "Connecticut", "DE": "Delaware",
		"FL": "Florida", "GA": "Georgia", "HI": "Hawaii", "ID": "Idaho",
		"IL": "Illinois", "IN": "Indiana", "IA": "Iowa", "KS": "Kansas",
		"KY": "Kentucky", "LA": "Louisiana", "ME": "Maine", "MD": "Maryland",
		"MA": "Massachusetts", "MI": "Michigan", "MN": "Minnesota", "MS": "Mississippi",
		"MO": "Missouri", "MT": "Montana", "NE": "Nebraska", "NV": "Nevada",
		"NH": "New Hampshire", "NJ": "New Jersey", "NM": "New Mexico", "NY": "New York",
		"NC": "North Carolina", "ND": "North Dakota", "OH": "Ohio", "OK": "Oklahoma",
		"OR": "Oregon", "PA": "Pennsylvania", "RI": "Rhode Island", "SC": "South Carolina",
		"SD": "South Dakota", "TN": "Tennessee", "TX": "Texas", "UT": "Utah",
		"VT": "Vermont", "VA": "Virginia", "WA": "Washington", "WV": "West Virginia",
		"WI": "Wisconsin", "WY": "Wyoming", "DC": "District of Columbia",
	},
	"CA": {
		"AB": "Alberta", "BC": "British Columbia", "MB": "Manitoba",
		"NB": "New Brunswick", "NL": "Newfoundland and Labrador", "NS": "Nova Scotia",
		"NT": "Northwest Territories", "NU": "Nunavut", "ON": "Ontario",
		"PE": "Prince Edward Island", "QC": "Quebec", "SK": "Saskatchewan",
		"YT": "Yukon",
	},
}
