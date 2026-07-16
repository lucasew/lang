package language

// EnglishVariant ports English locale Language subclasses (metadata surface).
type EnglishVariant struct {
	ShortCode    string   // en-US
	Name         string   // English (US)
	Countries    []string // US
	SpellerRuleID string
	// RelevantExtraRuleIDs are locale-specific rules beyond base English.
	RelevantExtraRuleIDs []string
}

func (v EnglishVariant) GetShortCodeWithCountryAndVariant() string { return v.ShortCode }
func (v EnglishVariant) GetName() string                           { return v.Name }
func (v EnglishVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// Predefined English variants.
var (
	AmericanEnglish = EnglishVariant{
		ShortCode: "en-US", Name: "English (US)", Countries: []string{"US"},
		SpellerRuleID:        "MORFOLOGIK_RULE_EN_US",
		RelevantExtraRuleIDs: []string{"AMERICAN_ENGLISH_REPLACE", "UNIT_CONVERSION_RULE_US"},
	}
	BritishEnglish = EnglishVariant{
		ShortCode: "en-GB", Name: "English (British)", Countries: []string{"GB"},
		SpellerRuleID:        "MORFOLOGIK_RULE_EN_GB",
		RelevantExtraRuleIDs: []string{"BRITISH_ENGLISH_REPLACE"},
	}
	CanadianEnglish = EnglishVariant{
		ShortCode: "en-CA", Name: "English (Canadian)", Countries: []string{"CA"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_CA",
	}
	AustralianEnglish = EnglishVariant{
		ShortCode: "en-AU", Name: "English (Australian)", Countries: []string{"AU"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_AU",
	}
	NewZealandEnglish = EnglishVariant{
		ShortCode: "en-NZ", Name: "English (New Zealand)", Countries: []string{"NZ"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_NZ",
	}
	SouthAfricanEnglish = EnglishVariant{
		ShortCode: "en-ZA", Name: "English (South African)", Countries: []string{"ZA"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_ZA",
	}
)

// AllEnglishVariants lists supported English locales.
func AllEnglishVariants() []EnglishVariant {
	return []EnglishVariant{
		AmericanEnglish, BritishEnglish, CanadianEnglish,
		AustralianEnglish, NewZealandEnglish, SouthAfricanEnglish,
	}
}

// EnglishVariantByCode looks up by short code (case-insensitive).
func EnglishVariantByCode(code string) (EnglishVariant, bool) {
	for _, v := range AllEnglishVariants() {
		if equalFoldASCII(v.ShortCode, code) {
			return v, true
		}
	}
	return EnglishVariant{}, false
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
