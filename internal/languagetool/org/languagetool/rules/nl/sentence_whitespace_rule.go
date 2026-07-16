package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWhitespaceRule wraps the core SentenceWhitespaceRule for this language.
type SentenceWhitespaceRule struct {
	*rules.SentenceWhitespaceRule
}

func NewSentenceWhitespaceRule(messages map[string]string) *SentenceWhitespaceRule {
	base := rules.NewSentenceWhitespaceRule(messages)
	base.RuleID = "NL_SENTENCE_WHITESPACE"
	base.MessageAfterSentence = "Voeg een spatie tussen zinnen toe."
	base.MessageAfterNumber = "Voeg een spatie toe na rangtelwoorden."
	return &SentenceWhitespaceRule{SentenceWhitespaceRule: base}
}

func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.SentenceWhitespaceRule.MatchList(sentences)
}
