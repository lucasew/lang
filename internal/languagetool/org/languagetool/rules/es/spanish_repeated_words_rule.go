package es

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/synonyms.txt
var synonymsFS embed.FS

var (
	synOnce sync.Once
	synMap  map[string]*rules.SynonymsData
)

func loadSynonyms() map[string]*rules.SynonymsData {
	synOnce.Do(func() {
		f, err := synonymsFS.Open("data/synonyms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSynonymsWords(f)
		if err != nil {
			panic(err)
		}
		synMap = m
	})
	return synMap
}

// SpanishRepeatedWordsRule ports org.languagetool.rules.es.SpanishRepeatedWordsRule (surface lemmas).
type SpanishRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewSpanishRepeatedWordsRule(messages map[string]string) *SpanishRepeatedWordsRule {
	base := &rules.AbstractRepeatedWordsRule{
		Messages:     messages,
		ID:           "ES_REPEATEDWORDS",
		Description:  "Repeated words",
		Message:      "Repeated word — consider a synonym.",
		ShortMsg:     "Style: Repeated word",
		WordsToCheck: loadSynonyms(),
	}
	return &SpanishRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

func (r *SpanishRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}
