package it

import (
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
	if position <= 0 || tokens[position] == nil || tokens[position-1] == nil {
		return false
	}
	// Java WordRepeatRule.wordRepetitionOf uses Token.equals (case-sensitive).
	prev, cur := tokens[position-1].GetToken(), tokens[position].GetToken()
	for _, w := range []string{"così", "passo", "piano", "via"} {
		if prev == w && cur == w {
			return true
		}
	}
	return false
}
