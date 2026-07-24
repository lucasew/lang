package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RuleFilter ports org.languagetool.rules.patterns.RuleFilter helpers.
// Concrete filters implement AcceptRuleMatch.
type RuleFilter interface {
	AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
		patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch
}

// GetRequired returns map[key] or panics if missing.
func GetRequired(key string, m map[string]string) string {
	v, ok := m[key]
	if !ok {
		panic("Missing key '" + key + "'")
	}
	return v
}

// GetOptional returns map[key] or empty.
func GetOptional(key string, m map[string]string) string {
	return m[key]
}

// GetOptionalDefault returns map[key] or defaultValue.
func GetOptionalDefault(key string, m map[string]string, defaultValue string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}
