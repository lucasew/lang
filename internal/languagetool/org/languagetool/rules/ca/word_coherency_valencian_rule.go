package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/coherency-valencia.txt
var coherencyValFS embed.FS

var (
	coherencyValOnce sync.Once
	coherencyValData *rules.WordCoherencyData
)

func loadCoherencyVal() *rules.WordCoherencyData {
	coherencyValOnce.Do(func() {
		f, err := coherencyValFS.Open("data/coherency-valencia.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.LoadWordCoherencyData(f, "/ca/coherency-valencia.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyValData = d
	})
	return coherencyValData
}

// WordCoherencyValencianRule ports org.languagetool.rules.ca.WordCoherencyValencianRule.
type WordCoherencyValencianRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyValencianRule(messages map[string]string) *WordCoherencyValencianRule {
	d := loadCoherencyVal()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "CA_WORD_COHERENCY_VALENCIA",
		Description: "Detecta l'ús incoherent de diferents formes dins d'un text.",
		ShortMsg:    "Coherència",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "No és coherent usar '" + word1 + "' i '" + word2 + "' dins d'un mateix text."
		},
	}
	return &WordCoherencyValencianRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyValencianRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
