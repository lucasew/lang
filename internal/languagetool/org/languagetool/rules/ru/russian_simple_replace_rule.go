package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "RU_SIMPLE_REPLACE",
			Description:          "Поиск просторечий и ошибочных фраз",
			ShortMsg:             "Ошибка?",
			MessageTemplate:      "«$match» — просторечие, исправление: $suggestions",
			SuggestionsSeparator: " или ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ru",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ru/replace.txt"); err != nil {
			panic(err)
		}
		replaceBase = base
	})
	return replaceBase
}

// RussianSimpleReplaceRule ports org.languagetool.rules.ru.RussianSimpleReplaceRule.
type RussianSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewRussianSimpleReplaceRule(messages map[string]string) *RussianSimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	return &RussianSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

func (r *RussianSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
