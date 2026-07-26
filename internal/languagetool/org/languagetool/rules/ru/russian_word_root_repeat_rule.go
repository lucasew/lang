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
		// Java WordCoherencyDataLoader.loadWords("/ru/wordrootrep.txt"): exact semicolon
		// pairs only (no invent expansion). Lemmas come from the tagger at match time.
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
	// DefaultOff ports Rule.setDefaultOff() (Java constructor calls setDefaultOff()).
	DefaultOff bool
}

// NewRussianWordRootRepeatRule ports RussianWordRootRepeatRule(ResourceBundle messages).
func NewRussianWordRootRepeatRule(messages map[string]string) *RussianWordRootRepeatRule {
	d := loadWordRootData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "RU_WORD_ROOT_REPEAT",
		Description: "Повтор однокоренных слов",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		// Java getMessage: «w1» и «w2» – однокоренные слова, их не стоит использовать одновременно
		// (en-dash U+2013 between clause and rest).
		MessageFn: func(word1, word2 string) string {
			return "«" + word1 + "» и «" + word2 + "» – однокоренные слова, их не стоит использовать одновременно"
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java:
	// addExamplePair(Example.wrong("Абрикос рос в саду. У меня на столе стоит <marker>абрикосный</marker> сок."),
	//                Example.fixed("Абрикос рос в саду. У меня на столе стоит сок из <marker>абрикосов</marker>."));
	base.AddExamplePair(
		rules.Wrong("Абрикос рос в саду. У меня на столе стоит <marker>абрикосный</marker> сок."),
		rules.Fixed("Абрикос рос в саду. У меня на столе стоит сок из <marker>абрикосов</marker>."),
	)
	r := &RussianWordRootRepeatRule{AbstractWordCoherencyRule: base}
	// Java: setDefaultOff();
	r.SetDefaultOff()
	return r
}

// SetDefaultOff ports Rule.setDefaultOff().
func (r *RussianWordRootRepeatRule) SetDefaultOff() {
	if r != nil {
		r.DefaultOff = true
	}
}

// IsDefaultOff ports Rule.isDefaultOff().
func (r *RussianWordRootRepeatRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// MatchList ports AbstractWordCoherencyRule.match over sentences.
func (r *RussianWordRootRepeatRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
