package fr

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
	base.RuleID = "FR_SENTENCE_WHITESPACE"
	base.MessageAfterSentence = "Ajoutez un espace entre les phrases."
	base.MessageAfterNumber = "Ajoutez un espace après les nombres ordinaux."
	return &SentenceWhitespaceRule{SentenceWhitespaceRule: base}
}

func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.SentenceWhitespaceRule.MatchList(sentences)
}
