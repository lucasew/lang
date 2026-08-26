package fa

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
		// Java WordCoherencyDataLoader: file pairs only (no invent expand).
		// Inflected forms match via tagger lemmas in AbstractWordCoherencyRule;
		// Persian Language has no default tagger — surface forms from the file.
		d, err := rules.LoadWordCoherencyData(f, "/fa/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.fa.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "FA_WORD_COHERENCY",
		Description: "چند املا برای یک کلمه که یکی از آنها اولویت بیشتری دارد",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			// Java: "'" + word1 + "' و '" + word2 + "' نباید در یک جا استفاده شوند"
			return "'" + word1 + "' و '" + word2 + "' نباید در یک جا استفاده شوند"
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java: example pair is commented-out TODO only — do not invent examples.
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

// MatchList is the text-level entry point (AbstractWordCoherencyRule.match).
func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
