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
		// Java WordCoherencyDataLoader.loadWords: file pairs only (no invent suffixes).
		// Inflected forms (aufwendiger, …) match via tagger lemmas in AbstractWordCoherencyRule.
		d, err := rules.LoadWordCoherencyData(f, "/de/coherency.txt", false)
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
	// Java WordCoherencyRule + AbstractWordCoherencyRule: MISC category; no shortMessage override.
	base := &rules.AbstractWordCoherencyRule{
		ID:          "DE_WORD_COHERENCY",
		Description: "Einheitliche Schreibweise für Wörter mit mehr als einer korrekten Schreibweise",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "'" + word1 + "' und '" + word2 + "' sollten nicht gleichzeitig benutzt werden."
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java: Delphine → Delfine (consistent with first spelling)
	base.AddExamplePair(
		rules.Wrong("Die Delfine gehören zu den Zahnwalen. <marker>Delphine</marker> sind in allen Meeren verbreitet."),
		rules.Fixed("Die Delfine gehören zu den Zahnwalen. <marker>Delfine</marker> sind in allen Meeren verbreitet."),
	)
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}

// MinToCheckParagraph ports AbstractWordCoherencyRule.minToCheckParagraph (Java returns -1).
func (r *WordCoherencyRule) MinToCheckParagraph() int {
	return r.AbstractWordCoherencyRule.MinToCheckParagraph()
}
