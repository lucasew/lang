package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishUnpairedBracketsRule ports org.languagetool.rules.en.EnglishUnpairedBracketsRule
// (brackets only; quotes handled by EnglishUnpairedQuotesRule).
type EnglishUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

func NewEnglishUnpairedBracketsRule(messages map[string]string) *EnglishUnpairedBracketsRule {
	start := []string{"[", "(", "{"}
	end := []string{"]", ")", "}"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	return &EnglishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *EnglishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
