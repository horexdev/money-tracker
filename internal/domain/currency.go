package domain

import "strings"

// CurrencyInfo describes a single ISO 4217 currency.
type CurrencyInfo struct {
	Code   string
	Name   string
	Symbol string
}

// Currencies maps ISO 4217 codes to currency info.
// Covers all commonly traded and widely used currencies.
var Currencies = map[string]CurrencyInfo{
	"AED": {Code: "AED", Name: "UAE Dirham", Symbol: "د.إ"},
	"AFN": {Code: "AFN", Name: "Afghan Afghani", Symbol: "؋"},
	"ALL": {Code: "ALL", Name: "Albanian Lek", Symbol: "L"},
	"AMD": {Code: "AMD", Name: "Armenian Dram", Symbol: "֏"},
	"ANG": {Code: "ANG", Name: "Netherlands Antillean Guilder", Symbol: "ƒ"},
	"AOA": {Code: "AOA", Name: "Angolan Kwanza", Symbol: "Kz"},
	"ARS": {Code: "ARS", Name: "Argentine Peso", Symbol: "$"},
	"AUD": {Code: "AUD", Name: "Australian Dollar", Symbol: "A$"},
	"AWG": {Code: "AWG", Name: "Aruban Florin", Symbol: "ƒ"},
	"AZN": {Code: "AZN", Name: "Azerbaijani Manat", Symbol: "₼"},
	"BAM": {Code: "BAM", Name: "Bosnia-Herzegovina Convertible Mark", Symbol: "KM"},
	"BBD": {Code: "BBD", Name: "Barbadian Dollar", Symbol: "Bds$"},
	"BDT": {Code: "BDT", Name: "Bangladeshi Taka", Symbol: "৳"},
	"BGN": {Code: "BGN", Name: "Bulgarian Lev", Symbol: "лв"},
	"BHD": {Code: "BHD", Name: "Bahraini Dinar", Symbol: ".د.ب"},
	"BIF": {Code: "BIF", Name: "Burundian Franc", Symbol: "FBu"},
	"BMD": {Code: "BMD", Name: "Bermudian Dollar", Symbol: "$"},
	"BND": {Code: "BND", Name: "Brunei Dollar", Symbol: "B$"},
	"BOB": {Code: "BOB", Name: "Bolivian Boliviano", Symbol: "Bs."},
	"BRL": {Code: "BRL", Name: "Brazilian Real", Symbol: "R$"},
	"BSD": {Code: "BSD", Name: "Bahamian Dollar", Symbol: "$"},
	"BTN": {Code: "BTN", Name: "Bhutanese Ngultrum", Symbol: "Nu."},
	"BWP": {Code: "BWP", Name: "Botswana Pula", Symbol: "P"},
	"BYN": {Code: "BYN", Name: "Belarusian Ruble", Symbol: "Br"},
	"BZD": {Code: "BZD", Name: "Belize Dollar", Symbol: "BZ$"},
	"CAD": {Code: "CAD", Name: "Canadian Dollar", Symbol: "C$"},
	"CDF": {Code: "CDF", Name: "Congolese Franc", Symbol: "FC"},
	"CHF": {Code: "CHF", Name: "Swiss Franc", Symbol: "CHF"},
	"CLP": {Code: "CLP", Name: "Chilean Peso", Symbol: "$"},
	"CNY": {Code: "CNY", Name: "Chinese Yuan", Symbol: "¥"},
	"COP": {Code: "COP", Name: "Colombian Peso", Symbol: "$"},
	"CRC": {Code: "CRC", Name: "Costa Rican Colón", Symbol: "₡"},
	"CUP": {Code: "CUP", Name: "Cuban Peso", Symbol: "$"},
	"CVE": {Code: "CVE", Name: "Cape Verdean Escudo", Symbol: "$"},
	"CZK": {Code: "CZK", Name: "Czech Koruna", Symbol: "Kč"},
	"DJF": {Code: "DJF", Name: "Djiboutian Franc", Symbol: "Fdj"},
	"DKK": {Code: "DKK", Name: "Danish Krone", Symbol: "kr"},
	"DOP": {Code: "DOP", Name: "Dominican Peso", Symbol: "RD$"},
	"DZD": {Code: "DZD", Name: "Algerian Dinar", Symbol: "د.ج"},
	"EGP": {Code: "EGP", Name: "Egyptian Pound", Symbol: "E£"},
	"ERN": {Code: "ERN", Name: "Eritrean Nakfa", Symbol: "Nfk"},
	"ETB": {Code: "ETB", Name: "Ethiopian Birr", Symbol: "Br"},
	"EUR": {Code: "EUR", Name: "Euro", Symbol: "€"},
	"FJD": {Code: "FJD", Name: "Fijian Dollar", Symbol: "FJ$"},
	"FKP": {Code: "FKP", Name: "Falkland Islands Pound", Symbol: "£"},
	"GBP": {Code: "GBP", Name: "British Pound", Symbol: "£"},
	"GEL": {Code: "GEL", Name: "Georgian Lari", Symbol: "₾"},
	"GHS": {Code: "GHS", Name: "Ghanaian Cedi", Symbol: "GH₵"},
	"GIP": {Code: "GIP", Name: "Gibraltar Pound", Symbol: "£"},
	"GMD": {Code: "GMD", Name: "Gambian Dalasi", Symbol: "D"},
	"GNF": {Code: "GNF", Name: "Guinean Franc", Symbol: "FG"},
	"GTQ": {Code: "GTQ", Name: "Guatemalan Quetzal", Symbol: "Q"},
	"GYD": {Code: "GYD", Name: "Guyanese Dollar", Symbol: "G$"},
	"HKD": {Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$"},
	"HNL": {Code: "HNL", Name: "Honduran Lempira", Symbol: "L"},
	"HRK": {Code: "HRK", Name: "Croatian Kuna", Symbol: "kn"},
	"HTG": {Code: "HTG", Name: "Haitian Gourde", Symbol: "G"},
	"HUF": {Code: "HUF", Name: "Hungarian Forint", Symbol: "Ft"},
	"IDR": {Code: "IDR", Name: "Indonesian Rupiah", Symbol: "Rp"},
	"ILS": {Code: "ILS", Name: "Israeli Shekel", Symbol: "₪"},
	"INR": {Code: "INR", Name: "Indian Rupee", Symbol: "₹"},
	"IQD": {Code: "IQD", Name: "Iraqi Dinar", Symbol: "ع.د"},
	"IRR": {Code: "IRR", Name: "Iranian Rial", Symbol: "﷼"},
	"ISK": {Code: "ISK", Name: "Icelandic Króna", Symbol: "kr"},
	"JMD": {Code: "JMD", Name: "Jamaican Dollar", Symbol: "J$"},
	"JOD": {Code: "JOD", Name: "Jordanian Dinar", Symbol: "JD"},
	"JPY": {Code: "JPY", Name: "Japanese Yen", Symbol: "¥"},
	"KES": {Code: "KES", Name: "Kenyan Shilling", Symbol: "KSh"},
	"KGS": {Code: "KGS", Name: "Kyrgyzstani Som", Symbol: "сом"},
	"KHR": {Code: "KHR", Name: "Cambodian Riel", Symbol: "៛"},
	"KMF": {Code: "KMF", Name: "Comorian Franc", Symbol: "CF"},
	"KPW": {Code: "KPW", Name: "North Korean Won", Symbol: "₩"},
	"KRW": {Code: "KRW", Name: "South Korean Won", Symbol: "₩"},
	"KWD": {Code: "KWD", Name: "Kuwaiti Dinar", Symbol: "د.ك"},
	"KYD": {Code: "KYD", Name: "Cayman Islands Dollar", Symbol: "CI$"},
	"KZT": {Code: "KZT", Name: "Kazakhstani Tenge", Symbol: "₸"},
	"LAK": {Code: "LAK", Name: "Lao Kip", Symbol: "₭"},
	"LBP": {Code: "LBP", Name: "Lebanese Pound", Symbol: "ل.ل"},
	"LKR": {Code: "LKR", Name: "Sri Lankan Rupee", Symbol: "Rs"},
	"LRD": {Code: "LRD", Name: "Liberian Dollar", Symbol: "L$"},
	"LSL": {Code: "LSL", Name: "Lesotho Loti", Symbol: "L"},
	"LYD": {Code: "LYD", Name: "Libyan Dinar", Symbol: "ل.د"},
	"MAD": {Code: "MAD", Name: "Moroccan Dirham", Symbol: "MAD"},
	"MDL": {Code: "MDL", Name: "Moldovan Leu", Symbol: "L"},
	"MGA": {Code: "MGA", Name: "Malagasy Ariary", Symbol: "Ar"},
	"MKD": {Code: "MKD", Name: "Macedonian Denar", Symbol: "ден"},
	"MMK": {Code: "MMK", Name: "Myanmar Kyat", Symbol: "K"},
	"MNT": {Code: "MNT", Name: "Mongolian Tugrik", Symbol: "₮"},
	"MOP": {Code: "MOP", Name: "Macanese Pataca", Symbol: "MOP$"},
	"MRU": {Code: "MRU", Name: "Mauritanian Ouguiya", Symbol: "UM"},
	"MUR": {Code: "MUR", Name: "Mauritian Rupee", Symbol: "₨"},
	"MVR": {Code: "MVR", Name: "Maldivian Rufiyaa", Symbol: "Rf"},
	"MWK": {Code: "MWK", Name: "Malawian Kwacha", Symbol: "MK"},
	"MXN": {Code: "MXN", Name: "Mexican Peso", Symbol: "MX$"},
	"MYR": {Code: "MYR", Name: "Malaysian Ringgit", Symbol: "RM"},
	"MZN": {Code: "MZN", Name: "Mozambican Metical", Symbol: "MT"},
	"NAD": {Code: "NAD", Name: "Namibian Dollar", Symbol: "N$"},
	"NGN": {Code: "NGN", Name: "Nigerian Naira", Symbol: "₦"},
	"NIO": {Code: "NIO", Name: "Nicaraguan Córdoba", Symbol: "C$"},
	"NOK": {Code: "NOK", Name: "Norwegian Krone", Symbol: "kr"},
	"NPR": {Code: "NPR", Name: "Nepalese Rupee", Symbol: "₨"},
	"NZD": {Code: "NZD", Name: "New Zealand Dollar", Symbol: "NZ$"},
	"OMR": {Code: "OMR", Name: "Omani Rial", Symbol: "ر.ع."},
	"PAB": {Code: "PAB", Name: "Panamanian Balboa", Symbol: "B/."},
	"PEN": {Code: "PEN", Name: "Peruvian Sol", Symbol: "S/."},
	"PGK": {Code: "PGK", Name: "Papua New Guinean Kina", Symbol: "K"},
	"PHP": {Code: "PHP", Name: "Philippine Peso", Symbol: "₱"},
	"PKR": {Code: "PKR", Name: "Pakistani Rupee", Symbol: "₨"},
	"PLN": {Code: "PLN", Name: "Polish Zloty", Symbol: "zł"},
	"PYG": {Code: "PYG", Name: "Paraguayan Guarani", Symbol: "₲"},
	"QAR": {Code: "QAR", Name: "Qatari Riyal", Symbol: "ر.ق"},
	"RON": {Code: "RON", Name: "Romanian Leu", Symbol: "lei"},
	"RSD": {Code: "RSD", Name: "Serbian Dinar", Symbol: "din."},
	"RUB": {Code: "RUB", Name: "Russian Ruble", Symbol: "₽"},
	"RWF": {Code: "RWF", Name: "Rwandan Franc", Symbol: "RF"},
	"SAR": {Code: "SAR", Name: "Saudi Riyal", Symbol: "ر.س"},
	"SBD": {Code: "SBD", Name: "Solomon Islands Dollar", Symbol: "SI$"},
	"SCR": {Code: "SCR", Name: "Seychellois Rupee", Symbol: "₨"},
	"SDG": {Code: "SDG", Name: "Sudanese Pound", Symbol: "ج.س."},
	"SEK": {Code: "SEK", Name: "Swedish Krona", Symbol: "kr"},
	"SGD": {Code: "SGD", Name: "Singapore Dollar", Symbol: "S$"},
	"SHP": {Code: "SHP", Name: "Saint Helena Pound", Symbol: "£"},
	"SLE": {Code: "SLE", Name: "Sierra Leonean Leone", Symbol: "Le"},
	"SOS": {Code: "SOS", Name: "Somali Shilling", Symbol: "Sh"},
	"SRD": {Code: "SRD", Name: "Surinamese Dollar", Symbol: "SRD"},
	"SSP": {Code: "SSP", Name: "South Sudanese Pound", Symbol: "£"},
	"STN": {Code: "STN", Name: "São Tomé and Príncipe Dobra", Symbol: "Db"},
	"SYP": {Code: "SYP", Name: "Syrian Pound", Symbol: "£S"},
	"SZL": {Code: "SZL", Name: "Eswatini Lilangeni", Symbol: "E"},
	"THB": {Code: "THB", Name: "Thai Baht", Symbol: "฿"},
	"TJS": {Code: "TJS", Name: "Tajikistani Somoni", Symbol: "SM"},
	"TMT": {Code: "TMT", Name: "Turkmenistani Manat", Symbol: "T"},
	"TND": {Code: "TND", Name: "Tunisian Dinar", Symbol: "د.ت"},
	"TOP": {Code: "TOP", Name: "Tongan Pa'anga", Symbol: "T$"},
	"TRY": {Code: "TRY", Name: "Turkish Lira", Symbol: "₺"},
	"TTD": {Code: "TTD", Name: "Trinidad and Tobago Dollar", Symbol: "TT$"},
	"TWD": {Code: "TWD", Name: "New Taiwan Dollar", Symbol: "NT$"},
	"TZS": {Code: "TZS", Name: "Tanzanian Shilling", Symbol: "TSh"},
	"UAH": {Code: "UAH", Name: "Ukrainian Hryvnia", Symbol: "₴"},
	"UGX": {Code: "UGX", Name: "Ugandan Shilling", Symbol: "USh"},
	"USD": {Code: "USD", Name: "US Dollar", Symbol: "$"},
	"UYU": {Code: "UYU", Name: "Uruguayan Peso", Symbol: "$U"},
	"UZS": {Code: "UZS", Name: "Uzbekistani Som", Symbol: "сўм"},
	"VES": {Code: "VES", Name: "Venezuelan Bolívar", Symbol: "Bs.S"},
	"VND": {Code: "VND", Name: "Vietnamese Dong", Symbol: "₫"},
	"VUV": {Code: "VUV", Name: "Vanuatu Vatu", Symbol: "VT"},
	"WST": {Code: "WST", Name: "Samoan Tala", Symbol: "WS$"},
	"XAF": {Code: "XAF", Name: "Central African CFA Franc", Symbol: "FCFA"},
	"XCD": {Code: "XCD", Name: "East Caribbean Dollar", Symbol: "EC$"},
	"XOF": {Code: "XOF", Name: "West African CFA Franc", Symbol: "CFA"},
	"XPF": {Code: "XPF", Name: "CFP Franc", Symbol: "₣"},
	"YER": {Code: "YER", Name: "Yemeni Rial", Symbol: "﷼"},
	"ZAR": {Code: "ZAR", Name: "South African Rand", Symbol: "R"},
	"ZMW": {Code: "ZMW", Name: "Zambian Kwacha", Symbol: "ZK"},
	"ZWL": {Code: "ZWL", Name: "Zimbabwean Dollar", Symbol: "Z$"},
}

// SearchCurrencies performs a case-insensitive prefix search on currency code and name.
// Returns up to 5 matching results, sorted code-first then name.
func SearchCurrencies(query string) []CurrencyInfo {
	if query == "" {
		return nil
	}
	q := strings.ToUpper(strings.TrimSpace(query))
	ql := strings.ToLower(q)

	var results []CurrencyInfo
	// First pass: code prefix matches (highest priority).
	for _, c := range Currencies {
		if strings.HasPrefix(c.Code, q) {
			results = append(results, c)
			if len(results) >= 5 {
				return results
			}
		}
	}

	// Second pass: name contains (case-insensitive).
	for _, c := range Currencies {
		if strings.Contains(strings.ToLower(c.Name), ql) {
			// Skip if already matched by code.
			dup := false
			for _, r := range results {
				if r.Code == c.Code {
					dup = true
					break
				}
			}
			if !dup {
				results = append(results, c)
				if len(results) >= 5 {
					return results
				}
			}
		}
	}
	return results
}

// ValidCurrency returns true if the code is a known ISO 4217 currency.
func ValidCurrency(code string) bool {
	_, ok := Currencies[strings.ToUpper(code)]
	return ok
}
