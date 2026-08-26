package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ItalianWordRepeatRule ports org.languagetool.rules.it.ItalianWordRepeatRule.
// Thin WordRepeatRule subclass: getId override + ignore exceptions, then super.ignore.
type ItalianWordRepeatRule struct {
	*rules.WordRepeatRule
}

// NewItalianWordRepeatRule ports ItalianWordRepeatRule(ResourceBundle, Language).
// Language is unused beyond the Java WordRepeatRule super ctor signature.
func NewItalianWordRepeatRule(messages map[string]string) *ItalianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "ITALIAN_WORD_REPEAT_RULE"
	r := &ItalianWordRepeatRule{WordRepeatRule: base}
	// Java: ignore() IT arms then return super.ignore(...).
	// ExtraIgnore true short-circuits; false continues base name list (Duran/…).
	base.ExtraIgnore = r.itIgnore
	return r
}

// itIgnore ports ItalianWordRepeatRule.ignore body (IT arms only).
// Uses wordRepetitionOf semantics: Token.equals (case-sensitive), not EqualFold.
// super.ignore runs via WordRepeatRule.Ignore after ExtraIgnore returns false.
func (r *ItalianWordRepeatRule) itIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if wordRepetitionOfIT("così", tokens, position) {
		return true // "Mi è sembrato così così."
	}
	if wordRepetitionOfIT("passo", tokens, position) {
		return true // "che passo passo vanno verso la rovina"
	}
	if wordRepetitionOfIT("piano", tokens, position) {
		return true // "Piano, piano, signor podestà."
	}
	if wordRepetitionOfIT("via", tokens, position) {
		return true // "Il lessico si andava via via modificando."
	}
	return false
}

// wordRepetitionOfIT ports WordRepeatRule.wordRepetitionOf:
// position > 0 && tokens[position-1].getToken().equals(word) && tokens[position].getToken().equals(word)
func wordRepetitionOfIT(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}
