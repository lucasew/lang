package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicWordRepeatRule ports org.languagetool.rules.ar.ArabicWordRepeatRule.
// Thin WordRepeatRule subclass: getId override + ignore exceptions, then super.ignore.
type ArabicWordRepeatRule struct {
	*rules.WordRepeatRule
}

// NewArabicWordRepeatRule ports ArabicWordRepeatRule(ResourceBundle).
// Java: super(messages, new Arabic()); Language is unused beyond the WordRepeatRule super ctor.
func NewArabicWordRepeatRule(messages map[string]string) *ArabicWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "ARABIC_WORD_REPEAT_RULE"
	// Java: فقط فقط → فقط
	base.AddExamplePair(
		rules.Wrong("هذا <marker>فقط فقط</marker> مثال."),
		rules.Fixed("هذا <marker>فقط</marker> مثال."),
	)
	r := &ArabicWordRepeatRule{WordRepeatRule: base}
	// Java: ignore() AR arms then return super.ignore(...).
	// ExtraIgnore true short-circuits; false continues base name list (Duran/…).
	base.ExtraIgnore = r.arIgnore
	return r
}

// arIgnore ports ArabicWordRepeatRule.ignore body (AR arms only).
// Uses wordRepetitionOf semantics: Token.equals (case-sensitive), not EqualFold.
// super.ignore runs via WordRepeatRule.Ignore after ExtraIgnore returns false.
func (r *ArabicWordRepeatRule) arIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if wordRepetitionOfAR("خطوة", tokens, position) {
		return true // "نفذت التعليمات خطوة خطوة."
	}
	if wordRepetitionOfAR("رويدا", tokens, position) {
		return true
	}
	return false
}

// wordRepetitionOfAR ports WordRepeatRule.wordRepetitionOf:
// position > 0 && tokens[position-1].getToken().equals(word) && tokens[position].getToken().equals(word)
func wordRepetitionOfAR(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}
