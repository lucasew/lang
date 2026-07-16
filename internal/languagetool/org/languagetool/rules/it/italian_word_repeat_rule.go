package it

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ItalianWordRepeatRule ports org.languagetool.rules.it.ItalianWordRepeatRule.
type ItalianWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewItalianWordRepeatRule(messages map[string]string) *ItalianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "ITALIAN_WORD_REPEAT_RULE"
	r := &ItalianWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.itIgnore
	return r
}

func (r *ItalianWordRepeatRule) itIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	for _, w := range []string{"così", "passo", "piano", "via"} {
		if wordRep(tokens, position, w) {
			return true
		}
	}
	return false
}

func wordRep(tokens []*languagetool.AnalyzedTokenReadings, position int, word string) bool {
	return strings.EqualFold(tokens[position-1].GetToken(), word) &&
		strings.EqualFold(tokens[position].GetToken(), word)
}
