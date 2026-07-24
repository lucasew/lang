package uk

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UkrainianUppercaseSentenceStartRule ports UkrainianUppercaseSentenceStartRule.
type UkrainianUppercaseSentenceStartRule struct {
	*rules.UppercaseSentenceStartRule
}

var ukLowerLetter = regexp.MustCompile(`^[а-яіїєґ]$`)

func NewUkrainianUppercaseSentenceStartRule(messages map[string]string) *UkrainianUppercaseSentenceStartRule {
	base := rules.NewUppercaseSentenceStartRule(messages, "uk")
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		// list item: а) б) в) — Java uses getCleanToken()
		if tokenIdx == 1 && tokenIdx < len(tokens)-1 &&
			tokens[tokenIdx] != nil && tokens[tokenIdx+1] != nil &&
			ukLowerLetter.MatchString(tokens[tokenIdx].GetCleanToken()) &&
			tokens[tokenIdx+1].GetToken() == ")" {
			return true
		}
		return false
	}
	return &UkrainianUppercaseSentenceStartRule{UppercaseSentenceStartRule: base}
}
