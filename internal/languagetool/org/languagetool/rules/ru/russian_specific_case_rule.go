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
		// Java: JLanguageTool.getDataBroker().getFromResourceDirAsLines("/ru/specific_case.txt")
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
// Java extends Rule directly (not AbstractSpecificCaseRule); match algorithm is identical.
// Shared AbstractSpecificCaseRule base carries the algorithm; RU configures messages/meta.
type RussianSpecificCaseRule struct {
	*rules.AbstractSpecificCaseRule
}

func NewRussianSpecificCaseRule(messages map[string]string) *RussianSpecificCaseRule {
	m, maxLen := loadSpecificCase()
	base := &rules.AbstractSpecificCaseRule{
		Messages:     messages,
		LcToProper:   m,
		MaxPhraseLen: maxLen,
		ID:           "RU_SPECIFIC_CASE",
		// Java RussianSpecificCaseRule.getDescription
		Description: "Написание специальных наименований в верхнем или нижнем регистре",
		// Java match: allWordsUppercase(properSpelling) branch
		InitialCapitalMessage: "Для специальных наименований используйте начальную заглавную букву.",
		// Java match: else branch (full upstream string)
		OtherCapitalizationMessage: "Для специальных наименований используйте предложенное написание заглавных и строчных букв.",
		// Java: new RuleMatch(..., msg) — no short message argument and no getShortMessage override.
		// Do not invent a Russian short message; leave ShortMsg empty.
	}
	// Java: super.setCategory(Categories.CASING.getCategory(messages)); setLocQualityIssueType(Misspelling)
	rules.InitSpecificCaseMeta(base, messages)
	// Java addExamplePair: fixed example omits trailing period (upstream quirk)
	base.AddExamplePair(
		rules.Wrong("Река <marker>рытый банк</marker> находится в Прикаспийской низменности."),
		rules.Fixed("Река <marker>Рытый Банк</marker> находится в Прикаспийской низменности"),
	)
	return &RussianSpecificCaseRule{AbstractSpecificCaseRule: base}
}

func (r *RussianSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSpecificCaseRule.Match(sentence)
}
