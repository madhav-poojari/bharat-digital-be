package districts

// TODO check if the below is correct
// Minimal list of MH districts (hardcoded). Add all district codes/names as needed.
var MaharashtraDistricts = map[string]string{
	"1802": "THANE",
	"1803": "RAIGAD",
	"1804": "RATNAGIRI",
	"1805": "SINDHUDURG",
	"1806": "NASHIK",
	"1807": "DHULE",
	"1808": "JALGAON",
	"1809": "AHMEDNAGAR",
	"1810": "PUNE",
	"1811": "SATARA",
	"1812": "SANGLI",
	"1813": "SOLAPUR",
	"1814": "KOLHAPUR",
	"1815": "CHHATRAPATI SAMBHAJI NAGAR",
	"1816": "JALNA",
	"1818": "BEED",
	"1819": "NANDED",
	"1820": "DHARASHIV",
	"1821": "LATUR",
	"1822": "BULDHANA",
	"1824": "AMRAVATI",
	"1825": "YAVATMAL",
	"1826": "WARDHA",
	"1827": "NAGPUR",
	"1828": "BHANDARA",
	"1830": "GADCHIROLI",
	"1831": "NANDURBAR",
	"1832": "WASHIM",
	"1833": "GONDIA",
	"1834": "HINGOLI",
	"1835": "PALGHAR",
}

// helper to get keys in deterministic order if needed
var MHOrdered = []string{"1802", "1803", "1804", "1805", "1806", "1807", "1808", "1809", "1810", "1811", "1812", "1813", "1814", "1815", "1816", "1818", "1819", "1820", "1821", "1822", "1824", "1825", "1826", "1827", "1828", "1830", "1831", "1832", "1833", "1834", "1835"}
