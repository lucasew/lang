package de

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
		// Expand full forms (aufwendiger / aufwändigste) without a German tagger.
		d, err := rules.LoadWordCoherencyData(f, "/de/coherency.txt", true)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.de.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "DE_WORD_COHERENCY",
		Description: "Einheitliche Schreibweise für Wörter mit mehr als einer korrekten Schreibweise",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "'" + word1 + "' und '" + word2 + "' sollten nicht gleichzeitig benutzt werden."
		},
	}
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
