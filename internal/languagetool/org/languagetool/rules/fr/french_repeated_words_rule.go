package fr

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

// FrenchRepeatedWordsRule ports org.languagetool.rules.fr.FrenchRepeatedWordsRule (surface).
type FrenchRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewFrenchRepeatedWordsRule(messages map[string]string) *FrenchRepeatedWordsRule {
	base := &rules.AbstractRepeatedWordsRule{
		Messages:     messages,
		ID:           "FR_REPEATEDWORDS",
		Description:  "Synonymes de mots répétés.",
		Message:      "Ce mot apparaît déjà dans l'une des phrases précédant immédiatement celle-ci. Utilisez un synonyme pour apporter plus de variété à votre texte, excepté si la répétition est intentionnelle.",
		ShortMsg:     "Style : Mot répété",
		WordsToCheck: loadSynonyms(),
	}
	return &FrenchRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

func (r *FrenchRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}
