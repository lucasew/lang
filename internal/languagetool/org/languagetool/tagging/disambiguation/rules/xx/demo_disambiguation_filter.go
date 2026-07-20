package xx

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DemoDisambiguationFilter ports
// org.languagetool.tagging.disambiguation.rules.xx.DemoDisambiguationFilter
// (test-resources twin used by xx/disambiguation.xml).
// Keeps the match only when the first pattern token surface is "X9".
type DemoDisambiguationFilter struct{}

func NewDemoDisambiguationFilter() *DemoDisambiguationFilter {
	return &DemoDisambiguationFilter{}
}

func (DemoDisambiguationFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if len(patternTokens) > 0 && patternTokens[0] != nil && patternTokens[0].GetToken() == "X9" {
		return match
	}
	return nil
}

func init() {
	// Java RuleFilterCreator loads by fully-qualified class name.
	patterns.GlobalRuleFilterCreator.Register(
		"org.languagetool.tagging.disambiguation.rules.xx.DemoDisambiguationFilter",
		func() patterns.RuleFilter { return NewDemoDisambiguationFilter() },
	)
}
