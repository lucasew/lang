package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWithManRule is a surface stand-in for org.languagetool.rules.de.SentenceWithManRule.
// MinPercent 0 flags every sentence containing "man".
type SentenceWithManRule struct {
	Messages   map[string]string
	MinPercent int
}

func NewSentenceWithManRule(messages map[string]string) *SentenceWithManRule {
	return &SentenceWithManRule{Messages: messages, MinPercent: 0}
}

func (r *SentenceWithManRule) GetID() string { return "SENTENCE_WITH_MAN_DE" }

func (r *SentenceWithManRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.MinPercent != 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if strings.EqualFold(tokens[i].GetToken(), "man") {
			msg := "Sätze mit der indirekten Leseransprache 'man' sind stilistisch wenig elegant formuliert. Lässt sich das Wort vermeiden?"
			rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			rm.ShortMessage = "indirekte Anrede 'man'"
			return []*rules.RuleMatch{rm}
		}
	}
	return nil
}
