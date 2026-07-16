package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt
var compoundsFS embed.FS

var (
	compoundOnce sync.Once
	compoundData *rules.CompoundRuleData
)

func loadCompoundData() *rules.CompoundRuleData {
	compoundOnce.Do(func() {
		f, err := compoundsFS.Open("data/compounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.NewCompoundRuleData(f, "/ru/compounds.txt")
		if err != nil {
			panic(err)
		}
		compoundData = d
	})
	return compoundData
}

// RussianCompoundRule ports org.languagetool.rules.ru.RussianCompoundRule.
type RussianCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewRussianCompoundRule(messages map[string]string) *RussianCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "RU_COMPOUNDS",
		Description:                 "Правописание через дефис",
		WithHyphenMessage:           "Эти слова должны быть написаны через дефис.",
		WithoutHyphenMessage:        "Эти слова должны быть написаны слитно.",
		WithOrWithoutHyphenMessage:  "Эти слова могут быть написаны через дефис или слитно.",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCompoundData(),
	}
	return &RussianCompoundRule{AbstractCompoundRule: base}
}

func (r *RussianCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractCompoundRule.Match(sentence)
}
