package ca

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
	base.RuleID = "CA_SENTENCE_WHITESPACE"
	base.MessageAfterSentence = "Afegiu un espai entre oracions."
	base.MessageAfterNumber = "Afegiu un espai després dels nombres ordinals."
	return &SentenceWhitespaceRule{SentenceWhitespaceRule: base}
}

func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.SentenceWhitespaceRule.MatchList(sentences)
}
