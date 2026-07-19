package en

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
		// Java WordCoherencyDataLoader: file pairs only (no invent -ed/-ing expand).
		// Inflected forms match via tagger lemmas in AbstractWordCoherencyRule.
		d, err := rules.LoadWordCoherencyData(f, "/en/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.en.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "EN_WORD_COHERENCY",
		Description: "Coherent spelling of words with two admitted variants.",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "Do not mix variants of the same word ('" + word1 + "' and '" + word2 + "') within a single text."
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

// MatchList is the text-level entry point.
func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
