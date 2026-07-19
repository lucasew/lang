package sv

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
		d, err := rules.LoadWordCoherencyData(f, "/sv/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// WordCoherencyRule ports org.languagetool.rules.sv.WordCoherencyRule.
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "SV_WORD_COHERENCY",
		Description: "Konsekvent stavning av ord med flera korrekta former",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "Använd endast en av stavningsvarianterna '" + word1 + "' och '" + word2 + "' i en och samma text."
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java: multi-marker mejl/mail; first fixed marker = mejl
	base.AddExamplePair(
		rules.Wrong("Det är en blandning av <marker>mejl</marker> och <marker>mail</marker> i det du skriver."),
		rules.Fixed("Om du använder enbart <marker>mejl</marker> när du skriver <marker>mejl</marker> blir det mer konsekvent."),
	)
	return &WordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
