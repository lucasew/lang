package nl

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
		// No inflection expansion needed for NL twin (surface organogram/organigram).
		d, err := rules.LoadWordCoherencyData(f, "/nl/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.nl.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "NL_WORD_COHERENCY",
		Description: "Consistente spelling van woorden met meerdere correcte vormen.",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "Gebruik liever niet '" + word1 + "' en '" + word2 + "' door elkaar in een tekst."
		},
	}
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
