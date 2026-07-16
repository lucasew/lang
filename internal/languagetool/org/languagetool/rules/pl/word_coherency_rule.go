package pl

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
		// Expand case endings for full forms (blefu / bluffem).
		d, err := rules.LoadWordCoherencyData(f, "/pl/coherency.txt", true)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.pl.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "PL_WORD_COHERENCY",
		Description: "Jednolita pisownia wyrazów o obocznej dopuszczalnej pisowni",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "Formy „" + word1 + "” i „" + word2 + "” zwykle nie powinny być używane jednocześnie."
		},
	}
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
