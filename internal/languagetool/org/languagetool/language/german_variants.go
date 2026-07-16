package language

// GermanVariant ports German locale Language subclasses.
type GermanVariant struct {
	ShortCode            string
	Name                 string
	Countries            []string
	SpellerRuleID        string
	RelevantExtraRuleIDs []string
}

func (v GermanVariant) GetShortCodeWithCountryAndVariant() string { return v.ShortCode }
func (v GermanVariant) GetName() string                           { return v.Name }
func (v GermanVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

var (
	GermanyGerman = GermanVariant{
		ShortCode: "de-DE", Name: "German (Germany)", Countries: []string{"DE"},
		SpellerRuleID:        "GERMAN_SPELLER_RULE",
		RelevantExtraRuleIDs: []string{"GERMAN_COMPOUND", "CASE_RULE"},
	}
	AustrianGerman = GermanVariant{
		ShortCode: "de-AT", Name: "German (Austria)", Countries: []string{"AT"},
		SpellerRuleID: "AUSTRIAN_GERMAN_SPELLER_RULE",
	}
	SwissGerman = GermanVariant{
		ShortCode: "de-CH", Name: "German (Swiss)", Countries: []string{"CH"},
		SpellerRuleID: "SWISS_GERMAN_SPELLER_RULE",
	}
)

func AllGermanVariants() []GermanVariant {
	return []GermanVariant{GermanyGerman, AustrianGerman, SwissGerman}
}

func GermanVariantByCode(code string) (GermanVariant, bool) {
	for _, v := range AllGermanVariants() {
		if equalFoldASCII(v.ShortCode, code) {
			return v, true
		}
	}
	return GermanVariant{}, false
}
