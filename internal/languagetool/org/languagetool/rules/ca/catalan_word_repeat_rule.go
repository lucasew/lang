package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanWordRepeatRule ports org.languagetool.rules.ca.CatalanWordRepeatRule.
// Ignore is POS/lemma only (Java); without tagger readings those arms fail closed
// (no surface invent of _allow_repeat / LOC_ADV / multiword lemmas).
type CatalanWordRepeatRule struct {
	*rules.WordRepeatRule
}

func NewCatalanWordRepeatRule(messages map[string]string) *CatalanWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "CATALAN_WORD_REPEAT_RULE"
	r := &CatalanWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.caIgnore
	return r
}

// caIgnore ports CatalanWordRepeatRule.ignore; base WordRepeatRule.Ignore then applies super.
func (r *CatalanWordRepeatRule) caIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	cur := tokens[position]
	prev := tokens[position-1]
	if cur == nil || prev == nil {
		return false
	}
	// Java:
	// hasPosTag("_allow_repeat") on cur or prev
	// || hasPosTag("LOC_ADV") on cur
	// || hasLemma("Joan-Lluís Lluís") on cur
	// || hasLemma("Chitty Chitty Bang Bang") on cur
	if cur.HasPosTag("_allow_repeat") || prev.HasPosTag("_allow_repeat") ||
		cur.HasPosTag("LOC_ADV") ||
		cur.HasLemma("Joan-Lluís Lluís") ||
		cur.HasLemma("Chitty Chitty Bang Bang") {
		return true
	}
	return false
}
