package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishUnpairedBracketsRule ports org.languagetool.rules.en.EnglishUnpairedBracketsRule
// (brackets only; quotes handled by EnglishUnpairedQuotesRule since LT 6.4).
// Java has no active isNoException/preventMatch overrides (commented TODO-dead code).
type EnglishUnpairedBracketsRule struct {
	*rules.GenericUnpairedBracketsRule
}

// NewEnglishUnpairedBracketsRule ports
// EnglishUnpairedBracketsRule(ResourceBundle messages, Language language).
// Java's language argument is unused (not passed to super).
func NewEnglishUnpairedBracketsRule(messages map[string]string) *EnglishUnpairedBracketsRule {
	// Java EN_START_SYMBOLS / EN_END_SYMBOLS — quotes not in this rule.
	start := []string{"[", "(", "{"}
	end := []string{"]", ")", "}"}
	base := rules.NewGenericUnpairedBracketsRule(messages, start, end)
	// Java getId → EN_UNPAIRED_BRACKETS
	base.SetRuleID("EN_UNPAIRED_BRACKETS")
	// Java setUrl(Tools.getUrl("...#what-are-parentheses"))
	base.SetURL("https://languagetool.org/insights/post/punctuation-guide/#what-are-parentheses")
	// Java addExamplePair
	base.AddExamplePair(
		rules.Wrong("He lived in a <marker>(</marker>large house."),
		rules.Fixed("He lived in a <marker>(</marker>large<marker>)</marker> house."),
	)
	return &EnglishUnpairedBracketsRule{GenericUnpairedBracketsRule: base}
}

// MatchList ports TextLevelRule.match inherited from GenericUnpairedBracketsRule.
// Rebinds Rule on each match to this EN type so getId/getUrl match Java `this`.
func (r *EnglishUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.GenericUnpairedBracketsRule == nil {
		return nil
	}
	ms := r.GenericUnpairedBracketsRule.MatchList(sentences)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
