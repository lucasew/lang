package uk

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var ukListLetter = regexp.MustCompile(`^[а-яіїєґ]$`)

// UkrainianUppercaseSentenceStartRule ports org.languagetool.rules.uk.UkrainianUppercaseSentenceStartRule.
type UkrainianUppercaseSentenceStartRule struct {
	*rules.UppercaseSentenceStartRule
}

func NewUkrainianUppercaseSentenceStartRule(messages map[string]string) *UkrainianUppercaseSentenceStartRule {
	base := rules.NewUppercaseSentenceStartRule(messages, "uk")
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		// list items: а) б) в)
		if tokenIdx == 1 && tokenIdx < len(tokens)-1 &&
			ukListLetter.MatchString(tokens[tokenIdx].GetToken()) &&
			tokens[tokenIdx+1].GetToken() == ")" {
			return true
		}
		return false
	}
	return &UkrainianUppercaseSentenceStartRule{UppercaseSentenceStartRule: base}
}

func (r *UkrainianUppercaseSentenceStartRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.UppercaseSentenceStartRule.MatchList(sentences)
}
