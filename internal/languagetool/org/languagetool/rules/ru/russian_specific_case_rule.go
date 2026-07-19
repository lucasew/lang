package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/specific_case.txt
var specificCaseFS embed.FS

var (
	specificCaseOnce sync.Once
	specificCaseMap  map[string]string
	specificCaseMax  int
)

func loadSpecificCase() (map[string]string, int) {
	specificCaseOnce.Do(func() {
		f, err := specificCaseFS.Open("data/specific_case.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, maxLen, err := rules.LoadSpecificCasePhrases(f)
		if err != nil {
			panic(err)
		}
		specificCaseMap = m
		specificCaseMax = maxLen
	})
	return specificCaseMap, specificCaseMax
}

// RussianSpecificCaseRule ports org.languagetool.rules.ru.RussianSpecificCaseRule.
type RussianSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewRussianSpecificCaseRule(messages map[string]string) *RussianSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:                   messages,
		LcToProper:                 m,
		MaxPhraseLen:               maxLen,
		ID:                         "RU_SPECIFIC_CASE",
		Description:                "Написание специальных наименований в верхнем или нижнем регистре",
		InitialCapitalMessage:      "Для специальных наименований используйте начальную заглавную букву.",
		OtherCapitalizationMessage: "Для специальных наименований используйте предложенное написание заглавных букв.",
		ShortMsg:                   "Специальное написание",
	}
	// Java: рытый банк → Рытый Банк (fixed example omits trailing period, same as upstream)
	base.AddExamplePair(
		rules.Wrong("Река <marker>рытый банк</marker> находится в Прикаспийской низменности."),
		rules.Fixed("Река <marker>Рытый Банк</marker> находится в Прикаспийской низменности"),
	)
	return &RussianSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *RussianSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
