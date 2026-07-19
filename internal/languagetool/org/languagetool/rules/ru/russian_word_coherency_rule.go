package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/coherency.txt
var coherencyFS embed.FS

var (
	coherencyOnce sync.Once
	coherencyData *rules.WordCoherencyData
)

func loadCoherencyData() *rules.WordCoherencyData {
	coherencyOnce.Do(func() {
		f, err := coherencyFS.Open("data/coherency.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java WordCoherencyDataLoader: file pairs only (no invent case suffixes).
		// Inflected forms match via tagger lemmas in AbstractWordCoherencyRule.
		d, err := rules.LoadWordCoherencyData(f, "/ru/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// RussianWordCoherencyRule ports org.languagetool.rules.ru.RussianWordCoherencyRule.
type RussianWordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewRussianWordCoherencyRule(messages map[string]string) *RussianWordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "RU_WORD_COHERENCY",
		Description: "Единообразное написание слов с более чем одним допустимым написанием",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "«" + word1 + "» и «" + word2 + "» не следует использовать одновременно"
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	return &RussianWordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *RussianWordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
