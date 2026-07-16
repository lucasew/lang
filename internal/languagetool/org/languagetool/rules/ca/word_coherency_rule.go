package ca

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

func loadCoherency() *rules.WordCoherencyData {
	coherencyOnce.Do(func() {
		f, err := coherencyFS.Open("data/coherency.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.LoadWordCoherencyData(f, "/ca/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.ca.WordCoherencyRule
// (surface forms; no Catalan synthesizer for inflected replacements).
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherency()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "CA_WORD_COHERENCY",
		Description: "Detecta l'ús incoherent de diferents formes dins d'un text.",
		ShortMsg:    "Coherència",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "No és coherent usar '" + word1 + "' i '" + word2 + "' dins d'un mateix text."
		},
	}
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
