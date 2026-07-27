package ar

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

// loadCoherencyData ports Java static:
//
//	private static final Map<String, Set<String>> wordMap =
//	    new WordCoherencyDataLoader().loadWords("/ar/coherency.txt");
func loadCoherencyData() *rules.WordCoherencyData {
	coherencyOnce.Do(func() {
		f, err := coherencyFS.Open("data/coherency.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java WordCoherencyDataLoader: file pairs only (no invent expand).
		// Inflected / definite forms match via tagger lemmas in AbstractWordCoherencyRule
		// (e.g. الشؤون → lemma شؤون pairs with surface شئون).
		d, err := rules.LoadWordCoherencyData(f, "/ar/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// ArabicWordCoherencyRule ports org.languagetool.rules.ar.ArabicWordCoherencyRule.
type ArabicWordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

// NewArabicWordCoherencyRule ports ArabicWordCoherencyRule(ResourceBundle messages).
func NewArabicWordCoherencyRule(messages map[string]string) *ArabicWordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "AR_WORD_COHERENCY",
		Description: "ضبط انسجام التهجئة للكلمات التي تكتب بطرق مختلفة مقبولة.",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			// Java: "تجنب استعمال شكلين للكلمة نفسها ('" + word1 + "' و '" + word2 + "') في  النص نفسه."
			// Note: double space after "في" (Java source).
			return "تجنب استعمال شكلين للكلمة نفسها ('" + word1 + "' و '" + word2 + "') في  النص نفسه."
		},
	}
	// Java AbstractWordCoherencyRule: super.setCategory(Categories.MISC.getCategory(messages))
	rules.InitWordCoherencyMeta(base, messages)
	// Java addExamplePair (double space after "تهتم" matches Java source).
	base.AddExamplePair(
		rules.Wrong("وزارة الشؤون الخارجية تهتم  بكل <marker>شئون</marker> العالم."),
		rules.Fixed("وزارة الشؤون الخارجية تهتم  بكل <marker>شؤون</marker> العالم."),
	)
	return &ArabicWordCoherencyRule{AbstractWordCoherencyRule: base}
}

// MatchList is the text-level entry point (AbstractWordCoherencyRule.match).
func (r *ArabicWordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
