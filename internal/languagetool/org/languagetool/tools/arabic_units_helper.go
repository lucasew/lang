package tools

// ArabicUnitsHelper ports unit (tamyeez) agreement forms for numeric phrases.
var arabicUnitsMap = map[string]map[string]string{
	"دينار": {
		"feminin": "no",
		"one_raf3": "دينار", "one_nasb": "دينارًا", "one_jar": "دينارٍ",
		"two_raf3": "ديناران", "two_nasb": "دينارين", "two_jar": "دينارين",
		"plural_raf3": "دنانيرُ", "plural_nasb": "دنانيرَ", "plural_jar": "دنانيرَ",
	},
	"درهم": {
		"feminin": "no",
		"one_raf3": "درهم", "one_nasb": "درهمًا", "one_jar": "درهمٍ",
		"two_raf3": "درهمان", "two_nasb": "درهمين", "two_jar": "درهمين",
		"plural_raf3": "دراهمُ", "plural_nasb": "دراهمَ", "plural_jar": "دراهمَ",
	},
	"دولار": {
		"feminin": "no",
		"one_raf3": "دولار", "one_nasb": "دولارًا", "one_jar": "دولارٍ",
		"two_raf3": "دولاران", "two_nasb": "دولارين", "two_jar": "دولارين",
		"plural_raf3": "دولاراتٌ", "plural_nasb": "دولاراتٍ", "plural_jar": "دولاراتٍ",
	},
	"ليرة": {
		"feminin": "yes",
		"one_raf3": "ليرة", "one_nasb": "ليرةً", "one_jar": "ليرةٍ",
		"two_raf3": "ليرتان", "two_nasb": "ليرتين", "two_jar": "ليرتين",
		"plural_raf3": "ليراتٌ", "plural_nasb": "ليراتٍ", "plural_jar": "ليراتٍ",
	},
}

func IsArabicUnitFeminin(unit string) bool {
	e, ok := arabicUnitsMap[unit]
	return ok && e["feminin"] == "yes"
}

func IsArabicUnit(unit string) bool {
	_, ok := arabicUnitsMap[unit]
	return ok
}

func GetArabicUnitForm(unit, category, inflection string) string {
	if inflection == "" {
		inflection = "raf3"
	}
	key := category + "_" + inflection
	if e, ok := arabicUnitsMap[unit]; ok {
		if v, ok := e[key]; ok {
			return v
		}
		return "[" + unit + "]"
	}
	return "[[" + unit + "]]"
}

func GetArabicUnitOneForm(unit, inflection string) string {
	return GetArabicUnitForm(unit, "one", inflection)
}

func GetArabicUnitTwoForm(unit, inflection string) string {
	return GetArabicUnitForm(unit, "two", inflection)
}

func GetArabicUnitPluralForm(unit, inflection string) string {
	return GetArabicUnitForm(unit, "plural", inflection)
}
