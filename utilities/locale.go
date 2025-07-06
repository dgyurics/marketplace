package utilities

import (
	"regexp"
)

// Supported ISO 3166-1 alpha-2 countries
var SupportedCountries = map[string]bool{
	"AE": true,
	"AR": true,
	"AT": true,
	"AU": true,
	"BE": true,
	"BG": true,
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
	"RS": true,
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
	"BGN": true,
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
	"RSD": true,
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

var PostalCodePatterns = map[string]*regexp.Regexp{
	"AE": regexp.MustCompile(`.*`),                                    // UAE: Not mandatory
	"AR": regexp.MustCompile(`^([A-Z]\d{4}[A-Z]{3})|(\d{4})$`),        // Argentina C1425ABC or 1425
	"AT": regexp.MustCompile(`^\d{4}$`),                               // Austria 1234
	"AU": regexp.MustCompile(`^\d{4}$`),                               // Australia 4000
	"BE": regexp.MustCompile(`^\d{4}$`),                               // Belgium 1234
	"BG": regexp.MustCompile(`^\d{4}$`),                               // Bulgaria 1234
	"BR": regexp.MustCompile(`^\d{5}-\d{3}$`),                         // Brazil 12345-678
	"CA": regexp.MustCompile(`^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$`), // Canada A1A 1A1
	"CH": regexp.MustCompile(`^\d{4}$`),                               // Switzerland 1234
	"CI": regexp.MustCompile(`.*`),                                    // Côte d’Ivoire: no standard
	"CL": regexp.MustCompile(`^\d{7}$`),                               // Chile 8320000
	"CO": regexp.MustCompile(`^\d{6}$`),                               // Colombia 110111
	"CY": regexp.MustCompile(`^\d{4}$`),                               // Cyprus 1100
	"CZ": regexp.MustCompile(`^\d{3} ?\d{2}$`),                        // Czech Republic 110 00
	"DE": regexp.MustCompile(`^\d{5}$`),                               // Germany 12345
	"DK": regexp.MustCompile(`^\d{4}$`),                               // Denmark 1234
	"EE": regexp.MustCompile(`^\d{5}$`),                               // Estonia 12345
	"EG": regexp.MustCompile(`^\d{5}$`),                               // Egypt 12345
	"ES": regexp.MustCompile(`^\d{5}$`),                               // Spain 12345
	"FI": regexp.MustCompile(`^\d{5}$`),                               // Finland 12345
	"FR": regexp.MustCompile(`^\d{5}$`),                               // France 12345
	"GB": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]? ?\d[A-Z]{2}$`),    // United Kingdom SW1A 1AA
	"GH": regexp.MustCompile(`.*`),                                    // Ghana: optional (GhanaPost GPS flexible)
	"GI": regexp.MustCompile(`^GX11 1AA$`),                            // Gibraltar GX11 1AA
	"GR": regexp.MustCompile(`^\d{3} ?\d{2}$`),                        // Greece 123 45
	"HK": regexp.MustCompile(`.*`),                                    // Hong Kong: no postal codes
	"HR": regexp.MustCompile(`^\d{5}$`),                               // Croatia 12345
	"HU": regexp.MustCompile(`^\d{4}$`),                               // Hungary 1011
	"ID": regexp.MustCompile(`^\d{5}$`),                               // Indonesia 12345
	"IE": regexp.MustCompile(`^[A-Za-z0-9]{3} ?[A-Za-z0-9]{4}$`),      // Ireland D02 X285
	"IL": regexp.MustCompile(`^\d{7}$`),                               // Israel 6100001
	"IN": regexp.MustCompile(`^\d{6}$`),                               // India 110001
	"IS": regexp.MustCompile(`^\d{3}$`),                               // Iceland 123
	"IT": regexp.MustCompile(`^\d{5}$`),                               // Italy 12345
	"JP": regexp.MustCompile(`^\d{3}-\d{4}$`),                         // Japan 123-4567
	"KE": regexp.MustCompile(`^\d{5}$`),                               // Kenya 00100
	"KR": regexp.MustCompile(`^\d{5}$`),                               // South Korea 12345
	"LI": regexp.MustCompile(`^\d{4}$`),                               // Liechtenstein 9490
	"LK": regexp.MustCompile(`^\d{5}$`),                               // Sri Lanka (placeholder)
	"LT": regexp.MustCompile(`^\d{5}$`),                               // Lithuania 12345
	"LU": regexp.MustCompile(`^\d{4}$`),                               // Luxembourg 1234
	"LV": regexp.MustCompile(`^\d{4}$`),                               // Latvia 1234
	"MA": regexp.MustCompile(`^\d{5}$`),                               // Morocco 10000
	"MT": regexp.MustCompile(`^[A-Z]{3} ?\d{4}$`),                     // Malta MLA 1001
	"MX": regexp.MustCompile(`^\d{5}$`),                               // Mexico 12345
	"MY": regexp.MustCompile(`^\d{5}$`),                               // Malaysia 43000
	"NG": regexp.MustCompile(`^\d{6}$`),                               // Nigeria 100001
	"NL": regexp.MustCompile(`^\d{4}\s?[A-Z]{2}$`),                    // Netherlands 1234 AB
	"NO": regexp.MustCompile(`^\d{4}$`),                               // Norway 1234
	"NZ": regexp.MustCompile(`^\d{4}$`),                               // New Zealand 6011
	"PA": regexp.MustCompile(`^\d{4}$|^.+$`),                          // Panama: 4 digits or optional
	"PE": regexp.MustCompile(`^\d{5}$`),                               // Peru 15001
	"PH": regexp.MustCompile(`^\d{4}$`),                               // Philippines 1000
	"PL": regexp.MustCompile(`^\d{2}-\d{3}$`),                         // Poland 12-345
	"PT": regexp.MustCompile(`^\d{4}-\d{3}$`),                         // Portugal 1234-567
	"RO": regexp.MustCompile(`^\d{6}$`),                               // Romania 123456
	"RS": regexp.MustCompile(`^\d{5}$`),                               // Serbia 11000
	"SA": regexp.MustCompile(`^\d{5}$`),                               // Saudi Arabia (example: 11564)
	"SE": regexp.MustCompile(`^\d{3} ?\d{2}$`),                        // Sweden 123 45
	"SG": regexp.MustCompile(`^\d{6}$`),                               // Singapore 560123
	"SI": regexp.MustCompile(`^\d{4}$`),                               // Slovenia 1234
	"SK": regexp.MustCompile(`^\d{3} ?\d{2}$`),                        // Slovakia 123 45
	"TH": regexp.MustCompile(`^\d{5}$`),                               // Thailand 10110
	"TW": regexp.MustCompile(`^\d{3}(-\d{2})?$`),                      // Taiwan 123 or 123-45
	"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),                      // United States 12345 or 12345-6789
	"UY": regexp.MustCompile(`^\d{5}$`),                               // Uruguay 11300
	"VN": regexp.MustCompile(`^\d{6}$`),                               // Vietnam 700000
	"ZA": regexp.MustCompile(`^\d{4}$`),                               // South Africa 2000
}
