package common

// ActiveProbesTable stores the active probe metadata
const ActiveProbesTable string = "active_probes"

// ProbeUpdatesTable stores every update message from probes
const ProbeUpdatesTable string = "probe_updates"

// JobsTable stores the jobs scheduled
const JobsTable string = "jobs"

// JobAlertsTable stores metadata about alert type jobs
const JobAlertsTable string = "job_alerts"

// JobTasksTable stores metadata abou task type jobs
const JobTasksTable string = "job_tasks"

// TasksTable stores metadata about task
const TasksTable string = "tasks"

// AccountsTable stores account information
const AccountsTable string = "accounts"

// CollectorsTable stores collector information
const CollectorsTable string = "collectors"

// TestHelpersTable stores the test helper information
const TestHelpersTable string = "test_helpers"

// URLsTable stores URL information
const URLsTable string = "urls"

// CountriesTable stores country information
const CountriesTable string = "countries"

// URLCategoriesTable stores the URL category code information
const URLCategoriesTable string = "url_categories"

type mapStrStruct map[string]struct{}

// AllCategoryCodes The list of all the supported category codes from the
// citizenlab test-list.
// See: https://github.com/citizenlab/test-lists/blob/master/lists/00-LEGEND-new_category_codes.csv
var AllCategoryCodes = mapStrStruct{"ALDR": {}, "REL": {}, "PORN": {},
	"PROV": {}, "POLR": {}, "HUMR": {}, "ENV": {}, "MILX": {},
	"HATE": {}, "NEWS": {}, "XED": {}, "PUBH": {}, "GMB": {},
	"ANON": {}, "DATE": {}, "GRP": {}, "LGBT": {}, "FILE": {},
	"HACK": {}, "COMT": {}, "MMED": {}, "HOST": {}, "SRCH": {},
	"GAME": {}, "CULTR": {}, "ECON": {}, "GOVT": {}, "COMM": {},
	"CTRL": {}, "IGO": {}, "MISC": {}}

// AllCountryCodes Official ISO 3166-2 alpha 2 country codes
// (https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) We also use the
// user-assignable ZZ and XX country codes to indicate "Unknown country" and
// "All countries" respectively .
var AllCountryCodes = mapStrStruct{
	"AD": {}, "AE": {}, "AF": {}, "AG": {}, "AI": {}, "AL": {},
	"AM": {}, "AN": {}, "AO": {}, "AQ": {}, "AR": {}, "AS": {},
	"AT": {}, "AU": {}, "AW": {}, "AZ": {}, "BA": {}, "BB": {},
	"BD": {}, "BE": {}, "BF": {}, "BG": {}, "BH": {}, "BI": {},
	"BJ": {}, "BM": {}, "BN": {}, "BO": {}, "BR": {}, "BS": {},
	"BT": {}, "BU": {}, "BV": {}, "BW": {}, "BY": {}, "BZ": {},
	"CA": {}, "CC": {}, "CF": {}, "CG": {}, "CH": {}, "CI": {},
	"CK": {}, "CL": {}, "CM": {}, "CN": {}, "CO": {}, "CR": {},
	"CS": {}, "CU": {}, "CV": {}, "CX": {}, "CY": {}, "CZ": {},
	"DD": {}, "DE": {}, "DJ": {}, "DK": {}, "DM": {}, "DO": {},
	"DZ": {}, "EC": {}, "EE": {}, "EG": {}, "EH": {}, "ER": {},
	"ES": {}, "ET": {}, "FI": {}, "FJ": {}, "FK": {}, "FM": {},
	"FO": {}, "FR": {}, "FX": {}, "GA": {}, "GB": {}, "GD": {},
	"GE": {}, "GF": {}, "GH": {}, "GI": {}, "GL": {}, "GM": {},
	"GN": {}, "GP": {}, "GQ": {}, "GR": {}, "GS": {}, "GT": {},
	"GU": {}, "GW": {}, "GY": {}, "HK": {}, "HM": {}, "HN": {},
	"HR": {}, "HT": {}, "HU": {}, "ID": {}, "IE": {}, "IL": {},
	"IN": {}, "IO": {}, "IQ": {}, "IR": {}, "IS": {}, "IT": {},
	"JM": {}, "JO": {}, "JP": {}, "KE": {}, "KG": {}, "KH": {},
	"KI": {}, "KM": {}, "KN": {}, "KP": {}, "KR": {}, "KW": {},
	"KY": {}, "KZ": {}, "LA": {}, "LB": {}, "LC": {}, "LI": {},
	"LK": {}, "LR": {}, "LS": {}, "LT": {}, "LU": {}, "LV": {},
	"LY": {}, "MA": {}, "MC": {}, "MD": {}, "MG": {}, "MH": {},
	"ML": {}, "MN": {}, "MM": {}, "MO": {}, "MP": {}, "MQ": {},
	"MR": {}, "MS": {}, "MT": {}, "MU": {}, "MV": {}, "MW": {},
	"MX": {}, "MY": {}, "MZ": {}, "NA": {}, "NC": {}, "NE": {},
	"NF": {}, "NG": {}, "NI": {}, "NL": {}, "NO": {}, "NP": {},
	"NR": {}, "NT": {}, "NU": {}, "NZ": {}, "OM": {}, "PA": {},
	"PE": {}, "PF": {}, "PG": {}, "PH": {}, "PK": {}, "PL": {},
	"PM": {}, "PN": {}, "PR": {}, "PT": {}, "PW": {}, "PY": {},
	"QA": {}, "RE": {}, "RO": {}, "RU": {}, "RW": {}, "SA": {},
	"SB": {}, "SC": {}, "SD": {}, "SE": {}, "SG": {}, "SH": {},
	"SI": {}, "SJ": {}, "SK": {}, "SL": {}, "SM": {}, "SN": {},
	"SO": {}, "SR": {}, "ST": {}, "SU": {}, "SV": {}, "SY": {},
	"SZ": {}, "TC": {}, "TD": {}, "TF": {}, "TG": {}, "TH": {},
	"TJ": {}, "TK": {}, "TM": {}, "TN": {}, "TO": {}, "TP": {},
	"TR": {}, "TT": {}, "TV": {}, "TW": {}, "TZ": {}, "UA": {},
	"UG": {}, "UM": {}, "US": {}, "UY": {}, "UZ": {}, "VA": {},
	"VC": {}, "VE": {}, "VG": {}, "VI": {}, "VN": {}, "VU": {},
	"WF": {}, "WS": {}, "YD": {}, "YE": {}, "YT": {}, "YU": {},
	"ZA": {}, "ZM": {}, "ZR": {}, "ZW": {}, "XX": {}}
