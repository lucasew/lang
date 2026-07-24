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
	// Java EnglishUnpairedBracketsRule.getId
	base.SetRuleID("EN_UNPAIRED_BRACKETS")
	// Java setUrl parentheses insights; addExamplePair
	base.URL = "https://languagetool.org/insights/post/punctuation-guide/#what-are-parentheses"
	// Java fixed example has two markers; correction is first marker span content only.
	base.AddExamplePair(
		rules.Wrong("He lived in a <marker>(</marker>large house."),
		rules.Fixed("He lived in a <marker>(</marker>large<marker>)</marker> house."),
	)
	return &EnglishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

func (r *EnglishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GenericUnpairedBracketsRule.MatchList(sentences)
}
