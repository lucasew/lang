package ru

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wordrootrep.txt
var wordRootFS embed.FS

var (
	wordRootOnce sync.Once
	wordRootData *rules.WordCoherencyData
)

func loadWordRootData() *rules.WordCoherencyData {
	wordRootOnce.Do(func() {
		f, err := wordRootFS.Open("data/wordrootrep.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Surface forms only (no Russian tagger expansion).
		d, err := rules.LoadWordCoherencyData(f, "/ru/wordrootrep.txt", false)
		if err != nil {
			panic(err)
		}
		wordRootData = d
	})
	return wordRootData
}

// RussianWordRootRepeatRule ports org.languagetool.rules.ru.RussianWordRootRepeatRule.
type RussianWordRootRepeatRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewRussianWordRootRepeatRule(messages map[string]string) *RussianWordRootRepeatRule {
	d := loadWordRootData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "RU_WORD_ROOT_REPEAT",
		Description: "Повтор однокоренных слов",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "«" + word1 + "» и «" + word2 + "» – однокоренные слова, их не стоит использовать одновременно"
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	return &RussianWordRootRepeatRule{AbstractWordCoherencyRule: base}
}

func (r *RussianWordRootRepeatRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
