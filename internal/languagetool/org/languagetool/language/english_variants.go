package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// EnglishVariant ports English locale Language subclasses (metadata surface).
type EnglishVariant struct {
	ShortCode    string   // en-US
	Name         string   // English (US)
	Countries    []string // US
	SpellerRuleID string
	// RelevantExtraRuleIDs are locale-specific rules beyond base English.
	RelevantExtraRuleIDs []string
}

// GetShortCode ports English.getShortCode ("en" for all locales).
func (v EnglishVariant) GetShortCode() string { return "en" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Stored ShortCode is the full locale code (en-US, en-GB, …).
func (v EnglishVariant) GetShortCodeWithCountryAndVariant() string { return v.ShortCode }

func (v EnglishVariant) GetName() string { return v.Name }
func (v EnglishVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → en/common_words.txt.
func (v EnglishVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// Predefined English variants.
var (
	AmericanEnglish = EnglishVariant{
		ShortCode: "en-US", Name: "English (US)", Countries: []string{"US"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_US",
		// Java AmericanEnglish.getRelevantRules: AmericanReplaceRule + UnitConversionRuleUS.
		RelevantExtraRuleIDs: []string{"EN_US_SIMPLE_REPLACE", "METRIC_UNITS_EN_US"},
	}
	BritishEnglish = EnglishVariant{
		ShortCode: "en-GB", Name: "English (British)", Countries: []string{"GB"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_GB",
		// Java BritishEnglish.getRelevantRules: BritishReplaceRule + UnitConversionRuleImperial.
		RelevantExtraRuleIDs: []string{"EN_GB_SIMPLE_REPLACE", "METRIC_UNITS_EN_IMPERIAL"},
	}
	CanadianEnglish = EnglishVariant{
		ShortCode: "en-CA", Name: "English (Canadian)", Countries: []string{"CA"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_CA",
		// Java CanadianEnglish: UnitConversionRuleImperial only.
		RelevantExtraRuleIDs: []string{"METRIC_UNITS_EN_IMPERIAL"},
	}
	AustralianEnglish = EnglishVariant{
		ShortCode: "en-AU", Name: "English (Australian)", Countries: []string{"AU"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_AU",
		// Java AustralianEnglish: UnitConversionRuleImperial only.
		RelevantExtraRuleIDs: []string{"METRIC_UNITS_EN_IMPERIAL"},
	}
	NewZealandEnglish = EnglishVariant{
		ShortCode: "en-NZ", Name: "English (New Zealand)", Countries: []string{"NZ"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_NZ",
		// Java NewZealandEnglish: NewZealandReplaceRule + UnitConversionRuleImperial.
		RelevantExtraRuleIDs: []string{"EN_NZ_SIMPLE_REPLACE", "METRIC_UNITS_EN_IMPERIAL"},
	}
	SouthAfricanEnglish = EnglishVariant{
		ShortCode: "en-ZA", Name: "English (South African)", Countries: []string{"ZA"},
		SpellerRuleID: "MORFOLOGIK_RULE_EN_ZA",
		// Java SouthAfricanEnglish: super only — no extras.
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

// GetMaintainedState ports English.getMaintainedState → ActivelyMaintained.
func (v EnglishVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports English.getMaintainers (Mike Unwalla, Marcin Miłkowski, Daniel Naber).
func (v EnglishVariant) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Mike Unwalla"),
		MarcinMilkowski,
		DanielNaber,
	}
}

// HasNGramFalseFriendRule ports English.hasNGramFalseFriendRule for mother short code.
func (v EnglishVariant) HasNGramFalseFriendRule(motherTongueShortCode string) bool {
	return EnglishHasNGramFalseFriendRule(motherTongueShortCode)
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / Morfologik* getId.
func (v EnglishVariant) GetDefaultSpellingRuleID() string {
	return v.SpellerRuleID
}

// GetOpeningDoubleQuote ports English.getOpeningDoubleQuote ("“").
func (v EnglishVariant) GetOpeningDoubleQuote() string { return "“" }

// GetClosingDoubleQuote ports English.getClosingDoubleQuote ("”").
func (v EnglishVariant) GetClosingDoubleQuote() string { return "”" }

// GetOpeningSingleQuote ports English.getOpeningSingleQuote ("‘").
func (v EnglishVariant) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports English.getClosingSingleQuote ("’").
func (v EnglishVariant) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports English.isAdvancedTypographyEnabled (true).
func (v EnglishVariant) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports English base Language.toAdvancedTypography with English quotes.
func (v EnglishVariant) ToAdvancedTypography(input string) string {
	return EnglishAdvancedTypography(input)
}

// HasMinMatchesRules ports English.hasMinMatchesRules (true).
func (v EnglishVariant) HasMinMatchesRules() bool { return true }

// GetDefaultRulePriorityForStyle ports English.getDefaultRulePriorityForStyle (-50).
func (v EnglishVariant) GetDefaultRulePriorityForStyle() int { return -50 }
