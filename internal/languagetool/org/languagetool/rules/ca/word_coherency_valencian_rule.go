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
// Same createReplacement synth path as WordCoherencyRule ([VAND].* + Synthesize).
type WordCoherencyValencianRule struct {
	*rules.AbstractWordCoherencyRule
	// Synthesize ports CatalanSynthesizer.synthesize(otherSpelling lemma, postag).
	Synthesize func(lemma, postag string) []string
}

func NewWordCoherencyValencianRule(messages map[string]string) *WordCoherencyValencianRule {
	d := loadCoherencyVal()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "CA_WORD_COHERENCY_VALENCIA",
		Description: "Detecta l'ús incoherent de diferents formes dins d'un text.",
		ShortMsg:    "Coherència",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		Category:    rules.CatStyle.GetCategory(messages),
		MessageFn: func(word1, word2 string) string {
			return "No és coherent usar '" + word1 + "' i '" + word2 + "' dins d'un mateix text."
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java multi-marker: Este… aquest → Este… este
	base.AddExamplePair(
		rules.Wrong("<marker>Este</marker> home d'ací parla amb <marker>aquest</marker> altre ací."),
		rules.Fixed("<marker>Este</marker> home d'ací parla amb <marker>este</marker> altre ací."),
	)
	r := &WordCoherencyValencianRule{AbstractWordCoherencyRule: base}
	base.CreateReplacement = r.createReplacement
	return r
}

func (r *WordCoherencyValencianRule) createReplacement(marked, token, otherSpelling string, atrs *languagetool.AnalyzedTokenReadings) string {
	if atrs != nil && r.Synthesize != nil {
		atr := atrs.ReadingWithTagRegex(caCoherencyAllowedPOS)
		if atr != nil && atr.GetPOSTag() != nil {
			forms := r.Synthesize(otherSpelling, *atr.GetPOSTag())
			if len(forms) > 0 {
				return forms[0]
			}
		}
	}
	return rules.DefaultWordCoherencyReplacement(marked, token, otherSpelling)
}

func (r *WordCoherencyValencianRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
