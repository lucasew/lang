package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PersianWordRepeatRule ports org.languagetool.rules.fa.PersianWordRepeatRule.
// Thin WordRepeatRule subclass: getId override + ignore exceptions, then super.ignore.
type PersianWordRepeatRule struct {
	*rules.WordRepeatRule
}

// NewPersianWordRepeatRule ports PersianWordRepeatRule(ResourceBundle, Language).
// Language is unused beyond the Java WordRepeatRule super ctor signature.
func NewPersianWordRepeatRule(messages map[string]string) *PersianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "PERSIAN_WORD_REPEAT_RULE"
	// Java: برای برای → برای
	base.AddExamplePair(
		rules.Wrong("این کار <marker>برای برای</marker> تو بود."),
		rules.Fixed("این کار <marker>برای</marker> تو بود."),
	)
	r := &PersianWordRepeatRule{WordRepeatRule: base}
	// Java: ignore() FA arms then return super.ignore(...).
	// ExtraIgnore true short-circuits; false continues base name list (Phi/Li/…).
	base.ExtraIgnore = r.faIgnore
	return r
}

// faIgnore ports PersianWordRepeatRule.ignore body (FA arms only).
// Uses wordRepetitionOf semantics: Token.equals (case-sensitive), not EqualFold.
// super.ignore runs via WordRepeatRule.Ignore after ExtraIgnore returns false.
func (r *PersianWordRepeatRule) faIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if wordRepetitionOfFA("لی", tokens, position) {
		return true // "من لی لی را دیدم"
	}
	if wordRepetitionOfFA("سی", tokens, position) {
		return true // "آب درون بطری ۱۰۰ سی سی بود"
	}
	if wordRepetitionOfFA("لک", tokens, position) {
		return true // "لک لک یک پرنده است"
	}
	if wordRepetitionOfFA("ریز", tokens, position) {
		return true // "غذایش را ریز ریز کرد"
	}
	if wordRepetitionOfFA("جز", tokens, position) {
		return true // "جز جز این کتاب را بلدم"
	}
	if wordRepetitionOfFA("کل", tokens, position) {
		return true // "با من کل کل نکن"
	}
	return false
}

// wordRepetitionOfFA ports WordRepeatRule.wordRepetitionOf:
// position > 0 && tokens[position-1].getToken().equals(word) && tokens[position].getToken().equals(word)
func wordRepetitionOfFA(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}
