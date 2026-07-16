package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishUnpairedQuotesRule ports org.languagetool.rules.en.EnglishUnpairedQuotesRule
// without POS-based apostrophe exceptions (surface AnalyzePlain only).
type EnglishUnpairedQuotesRule struct {
	*rules.GenericUnpairedQuotesRule
}

func NewEnglishUnpairedQuotesRule(messages map[string]string) *EnglishUnpairedQuotesRule {
	start := []string{"“", "\"", "'", "‘"}
	end := []string{"”", "\"", "'", "’"}
	base := rules.NewGenericUnpairedQuotesRule(messages, start, end)
	base.SetRuleID("EN_UNPAIRED_QUOTES")
	return &EnglishUnpairedQuotesRule{GenericUnpairedQuotesRule: base}
}

func (r *EnglishUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedQuotesRule.MatchList(sentences)
}
