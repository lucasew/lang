package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicWordRepeatRule ports org.languagetool.rules.ar.ArabicWordRepeatRule.
type ArabicWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewArabicWordRepeatRule(messages map[string]string) *ArabicWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "ARABIC_WORD_REPEAT_RULE"
	// Java: فقط فقط → فقط
	base.AddExamplePair(
		rules.Wrong("هذا <marker>فقط فقط</marker> مثال."),
		rules.Fixed("هذا <marker>فقط</marker> مثال."),
	)
	r := &ArabicWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.arIgnore
	return r
}

func (r *ArabicWordRepeatRule) arIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 || tokens[position] == nil || tokens[position-1] == nil {
		return false
	}
	// Java WordRepeatRule.wordRepetitionOf uses token equality (not fold).
	prev, cur := tokens[position-1].GetToken(), tokens[position].GetToken()
	for _, w := range []string{"خطوة", "رويدا"} {
		if prev == w && cur == w {
			return true
		}
	}
	return false
}
