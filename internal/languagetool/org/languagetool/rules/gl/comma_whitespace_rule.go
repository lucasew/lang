package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CommaWhitespaceRule wraps the core CommaWhitespaceRule for this language.
type CommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
}

func NewCommaWhitespaceRule(messages map[string]string) *CommaWhitespaceRule {
	return &CommaWhitespaceRule{CommaWhitespaceRule: rules.NewCommaWhitespaceRule(messages)}
}

func (r *CommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.CommaWhitespaceRule.Match(sentence)
}
