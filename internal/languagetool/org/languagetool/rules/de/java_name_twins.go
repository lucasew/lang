package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// GermanHelper is the Java-name twin for GermanHelper POS utilities.
type GermanHelper struct{}

func (GermanHelper) GetNounCase(posTag string) string   { return GetNounCase(posTag) }
func (GermanHelper) GetNounNumber(posTag string) string { return GetNounNumber(posTag) }
func (GermanHelper) GetNounGender(posTag string) string { return GetNounGender(posTag) }
func (GermanHelper) HasReadingOfType(tok *languagetool.AnalyzedTokenReadings, t POSType) bool {
	return HasReadingOfType(tok, t)
}

// GermanTools is the Java-name twin for GermanTools helpers.
type GermanTools struct{}

func (GermanTools) IsVowel(c rune) bool { return IsVowel(c) }

// PrepositionToCases is the Java-name twin for preposition→case government.
type PrepositionToCases struct{}

func (PrepositionToCases) Cases(prep string) []GrammaticalCase {
	return PrepositionCases[strings.ToLower(prep)]
}

func (PrepositionToCases) Allows(prep string, c GrammaticalCase) bool {
	for _, x := range PrepositionCases[strings.ToLower(prep)] {
		if x == c {
			return true
		}
	}
	return false
}

// CaseRuleAntiPatternsList wraps the CaseRule anti-pattern table (var CaseRuleAntiPatterns).
type CaseRuleAntiPatternsList struct{}

func (CaseRuleAntiPatternsList) All() [][]*patterns.PatternToken { return CaseRuleAntiPatterns }
func (CaseRuleAntiPatternsList) Count() int                      { return CaseRuleAntiPatternsCount() }

// CaseRuleExceptionsData wraps case-rule exception phrases.
type CaseRuleExceptionsData struct{}

func (CaseRuleExceptionsData) Contains(phrase string) bool { return IsCaseRuleException(phrase) }
func (CaseRuleExceptionsData) All() map[string]struct{}    { return CaseRuleExceptions() }

// AgreementRuleAntiPatterns1Data, AgreementRuleAntiPatterns2Data, AgreementRuleAntiPatterns3Data
// wrap the three anti-pattern tables (package vars share the Java names).
type AgreementRuleAntiPatterns1Data struct{}

func (AgreementRuleAntiPatterns1Data) Patterns() [][]*patterns.PatternToken {
	return AgreementRuleAntiPatterns1
}

type AgreementRuleAntiPatterns2Data struct{}

func (AgreementRuleAntiPatterns2Data) Patterns() [][]*patterns.PatternToken {
	return AgreementRuleAntiPatterns2
}

type AgreementRuleAntiPatterns3Data struct{}

func (AgreementRuleAntiPatterns3Data) Patterns() [][]*patterns.PatternToken {
	return AgreementRuleAntiPatterns3
}
