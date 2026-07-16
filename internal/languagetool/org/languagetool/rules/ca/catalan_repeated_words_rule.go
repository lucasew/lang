package ca

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

// CatalanRepeatedWordsRule ports org.languagetool.rules.ca.CatalanRepeatedWordsRule (surface lemmas).
type CatalanRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewCatalanRepeatedWordsRule(messages map[string]string) *CatalanRepeatedWordsRule {
	base := &rules.AbstractRepeatedWordsRule{
		Messages:     messages,
		ID:           "CA_REPEATEDWORDS",
		Description:  "Repeated words",
		Message:      "Repeated word — consider a synonym.",
		ShortMsg:     "Style: Repeated word",
		WordsToCheck: loadSynonyms(),
	}
	return &CatalanRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

func (r *CatalanRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}
