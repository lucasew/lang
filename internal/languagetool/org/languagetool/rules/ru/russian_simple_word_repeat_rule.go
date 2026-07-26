package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianSimpleWordRepeatRule ports org.languagetool.rules.ru.RussianSimpleWordRepeatRule.
// Java extends WordRepeatRule; ctor (messages, Language) — Language unused beyond super.
type RussianSimpleWordRepeatRule struct {
	*rules.WordRepeatRule
}

// PATTERN ports Java Pattern.compile("[a-zA-Zа-яёА-ЯЁ]"); matcher.matches() = full string.
var russianSimpleWordRepeatSingleLetter = regexp.MustCompile(`^[a-zA-Zа-яёА-ЯЁ]$`)

// NewRussianSimpleWordRepeatRule ports RussianSimpleWordRepeatRule(ResourceBundle, Language).
// Language is unused in Java beyond WordRepeatRule super (category/issue type from messages).
func NewRussianSimpleWordRepeatRule(messages map[string]string) *RussianSimpleWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	r := &RussianSimpleWordRepeatRule{WordRepeatRule: base}
	// ExtraIgnore = override ignore() body; WordRepeatRule.Ignore then runs super.ignore names.
	base.ExtraIgnore = r.ruIgnore
	return r
}

// ruIgnore ports RussianSimpleWordRepeatRule.ignore (then super via WordRepeatRule.Ignore).
// Order mirrors Java exactly.
func (r *RussianSimpleWordRepeatRule) ruIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if wordRepetitionOfRU("-", tokens, position) {
		return true
	}
	if wordRepetitionOfRU("и", tokens, position) {
		return true
	}
	if wordRepetitionOfRU("по", tokens, position) {
		return true
	}
	if position > 0 &&
		tokens[position-1].GetToken() == "ПО" && tokens[position].GetToken() == "по" {
		return true // "ПО по"
	}
	if position > 0 &&
		tokens[position-1].GetToken() == "по" && tokens[position].GetToken() == "ПО" {
		return true // "по ПО"
	}
	if wordRepetitionOfRU("что", tokens, position) {
		return true
	}
	if russianSimpleWordRepeatSingleLetter.MatchString(tokens[position].GetToken()) &&
		position > 1 &&
		russianSimpleWordRepeatSingleLetter.MatchString(tokens[position-1].GetToken()) {
		// spelling with spaces in between: "L L"
		return true
	}
	return false
}

// wordRepetitionOfRU ports WordRepeatRule.wordRepetitionOf:
// position > 0 && tokens[position-1].getToken().equals(word) && tokens[position].getToken().equals(word)
// (Java String.equals — case-sensitive; NOT equalsIgnoreCase / EqualFold).
func wordRepetitionOfRU(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		tokens[position-1].GetToken() == word &&
		tokens[position].GetToken() == word
}
