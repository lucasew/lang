package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianFillerWordsRule ports org.languagetool.rules.ru.RussianFillerWordsRule
// Java default minPercent=8 (AbstractFillerWordsRule).
type RussianFillerWordsRule struct {
	*rules.AbstractFillerWordsRule
}

func NewRussianFillerWordsRule(messages map[string]string) *RussianFillerWordsRule {
	fillers := map[string]struct{}{}
	for _, w := range []string{"ах", "аа", "ааа", "аааа", "ау", "бу", "вау", "ох", "однако", "эээ", "э", "эй", "эх", "ух-ты", "ух"} {
		fillers[w] = struct{}{}
	}
	base := &rules.AbstractFillerWordsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{},
		Messages:                   messages,
		ID:                         "FILLER_WORDS_RU",
		Description:                "Filler words",
		ShortMsg:                   "Filler word",
		Message:                    "Возможно, это слово-паразит.",
		FillerWords:                fillers,
	}
	rules.InitFillerWordsMeta(base, messages, false)
	return &RussianFillerWordsRule{AbstractFillerWordsRule: base}
}

func (r *RussianFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractFillerWordsRule.Match(sentence)
}
