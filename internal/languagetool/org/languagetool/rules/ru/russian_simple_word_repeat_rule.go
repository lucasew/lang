package ru

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianSimpleWordRepeatRule ports org.languagetool.rules.ru.RussianSimpleWordRepeatRule.
type RussianSimpleWordRepeatRule struct {
	*rules.WordRepeatRule
}

var singleLetterRU = regexp.MustCompile(`^[a-zA-Zа-яёА-ЯЁ]$`)

func NewRussianSimpleWordRepeatRule(messages map[string]string) *RussianSimpleWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	r := &RussianSimpleWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.ruIgnore
	return r
}

func (r *RussianSimpleWordRepeatRule) ruIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position <= 0 {
		return false
	}
	for _, w := range []string{"-", "и", "по", "что"} {
		if strings.EqualFold(tokens[position-1].GetToken(), w) && strings.EqualFold(tokens[position].GetToken(), w) {
			return true
		}
	}
	if tokens[position-1].GetToken() == "ПО" && tokens[position].GetToken() == "по" {
		return true
	}
	if tokens[position-1].GetToken() == "по" && tokens[position].GetToken() == "ПО" {
		return true
	}
	if singleLetterRU.MatchString(tokens[position].GetToken()) &&
		position > 1 &&
		singleLetterRU.MatchString(tokens[position-1].GetToken()) {
		return true
	}
	return false
}
