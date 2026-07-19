package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.ar.filters.ArabicDateCheckFilter", func() patterns.RuleFilter {
		return NewArabicDateCheckFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return arabicAdvancedSynthRuleFilter{}
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter", func() patterns.RuleFilter {
		return NewArabicNumberPhraseFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicMasdarToVerbFilter", func() patterns.RuleFilter {
		return NewArabicMasdarToVerbFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicAdjectiveToExclamationFilter", func() patterns.RuleFilter {
		return NewArabicAdjectiveToExclamationFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter", func() patterns.RuleFilter {
		return NewArabicVerbToMafoulMutlaqFilter()
	})
}

type arabicAdvancedSynthRuleFilter struct{}

func (arabicAdvancedSynthRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return (&rules.AbstractAdvancedSynthesizerFilter{}).AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
