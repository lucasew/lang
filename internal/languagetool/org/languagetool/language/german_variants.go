package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

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

// GetShortCode ports Language.getShortCode ("de" for all DE variants).
func (v GermanVariant) GetShortCode() string { return "de" }

// GetMaintainedState ports German.getMaintainedState → ActivelyMaintained.
func (v GermanVariant) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// IsVariant ports GermanyGerman/AustrianGerman/SwissGerman.isVariant() → true.
func (v GermanVariant) IsVariant() bool { return true }

// GetMaintainers ports German.getMaintainers (Jan Schreiber, Daniel Naber).
func (v GermanVariant) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Jan Schreiber"),
		DanielNaber,
	}
}

// GetOpeningDoubleQuote ports German.getOpeningDoubleQuote / SwissGerman override.
func (v GermanVariant) GetOpeningDoubleQuote() string {
	if isSwissGermanVariant(v) {
		return "«"
	}
	return "„"
}

// GetClosingDoubleQuote ports German.getClosingDoubleQuote / SwissGerman override.
func (v GermanVariant) GetClosingDoubleQuote() string {
	if isSwissGermanVariant(v) {
		return "»"
	}
	return "“"
}

// GetOpeningSingleQuote ports German.getOpeningSingleQuote ("‚").
func (v GermanVariant) GetOpeningSingleQuote() string { return "‚" }

// GetClosingSingleQuote ports German.getClosingSingleQuote ("‘").
func (v GermanVariant) GetClosingSingleQuote() string { return "‘" }

// GetIgnoredCharactersRegex ports German.getIgnoredCharactersRegex → [\u00AD].
func (v GermanVariant) GetIgnoredCharactersRegex() *regexp.Regexp {
	return languagetool.GermanIgnoredCharactersRegex
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / *GermanSpellerRule getId.
func (v GermanVariant) GetDefaultSpellingRuleID() string {
	return v.SpellerRuleID
}

// GetCommonWordsPath ports Language.getCommonWordsPath → de/common_words.txt.
func (v GermanVariant) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

func isSwissGermanVariant(v GermanVariant) bool {
	return strings.EqualFold(v.ShortCode, "de-CH") || strings.HasSuffix(strings.ToUpper(v.ShortCode), "-CH")
}

var (
	GermanyGerman = GermanVariant{
		ShortCode: "de-DE", Name: "German (Germany)", Countries: []string{"DE"},
		SpellerRuleID: "GERMAN_SPELLER_RULE",
		// Java GermanyGerman.getRelevantRules: super + GermanCompoundRule only.
		// DE_CASE is already in German.getRelevantRules (CaseRule).
		RelevantExtraRuleIDs: []string{"DE_COMPOUNDS"},
	}
	AustrianGerman = GermanVariant{
		ShortCode: "de-AT", Name: "German (Austria)", Countries: []string{"AT"},
		SpellerRuleID: "AUSTRIAN_GERMAN_SPELLER_RULE",
		// Java AustrianGerman.getRelevantRules: super + GermanCompoundRule (same as DE).
		RelevantExtraRuleIDs: []string{"DE_COMPOUNDS"},
	}
	SwissGerman = GermanVariant{
		ShortCode: "de-CH", Name: "German (Swiss)", Countries: []string{"CH"},
		SpellerRuleID: "SWISS_GERMAN_SPELLER_RULE",
	}
)

// deDEATGrammarXML is the shared DE/AT grammar path from NonSwissGerman /
// GermanyGerman / AustrianGerman.getRuleFileNames.
const deDEATGrammarXML = "/org/languagetool/rules/de/de-DE-AT/grammar.xml"

// GetRuleFileNames ports GermanyGerman / AustrianGerman / SwissGerman getRuleFileNames.
// Base Language list (always-true exists), then de-DE-AT grammar for non-Swiss.
func (v GermanVariant) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists is GetRuleFileNames with a pluggable style/variant exists probe.
func (v GermanVariant) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	out := languagetool.GetRuleFileNames("de", v.ShortCode, "/org/languagetool/rules", exists)
	if !isSwissGermanVariant(v) {
		// GermanyGerman, AustrianGerman, NonSwissGerman all append de-DE-AT/grammar.xml
		out = append(out, deDEATGrammarXML)
	}
	return out
}

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
