package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishWordRepeatRule ports org.languagetool.rules.es.SpanishWordRepeatRule.
// Ignore is POS-only (_allow_repeat); without tagger readings fail closed
// (no surface invent for /a a or .ES es).
type SpanishWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewSpanishWordRepeatRule(messages map[string]string) *SpanishWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "SPANISH_WORD_REPEAT_RULE"
	r := &SpanishWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.esIgnore
	return r
}

// esIgnore ports SpanishWordRepeatRule.ignore; base then applies super.ignore.
func (r *SpanishWordRepeatRule) esIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	cur, prev := tokens[position], tokens[position-1]
	if cur == nil || prev == nil {
		return false
	}
	// Java: hasPosTag("_allow_repeat") on cur or prev
	if cur.HasPosTag("_allow_repeat") || prev.HasPosTag("_allow_repeat") {
		return true
	}
	return false
}
