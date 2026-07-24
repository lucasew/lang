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

// caCoherencyAllowedPOS ports WordCoherencyRule.allowedPostags Pattern "[VAND].*"
const caCoherencyAllowedPOS = `[VAND].*`

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

// WordCoherencyRule ports org.languagetool.rules.ca.WordCoherencyRule.
// createReplacement synthesizes via Synthesize when a V/A/N/D POS reading exists;
// without Synthesize, falls back to surface default (Java always has CatalanSynthesizer).
type WordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
	// Synthesize ports CatalanSynthesizer.synthesize(token with lemma=otherSpelling, postag).
	Synthesize func(lemma, postag string) []string
}

func NewWordCoherencyRule(messages map[string]string) *WordCoherencyRule {
	d := loadCoherency()
	base := &rules.AbstractWordCoherencyRule{
		ID:          "CA_WORD_COHERENCY",
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
	// Java: pesebre / pessebre consistency (multi-marker wrong; first fixed marker = pesebre)
	base.AddExamplePair(
		rules.Wrong("Un <marker>pesebre</marker> ací i un altre <marker>pessebre</marker> allà."),
		rules.Fixed("Un <marker>pesebre</marker> ací i un altre <marker>pesebre</marker> allà."),
	)
	r := &WordCoherencyRule{AbstractWordCoherencyRule: base}
	base.CreateReplacement = r.createReplacement
	return r
}

// createReplacement ports WordCoherencyRule.createReplacement.
func (r *WordCoherencyRule) createReplacement(marked, token, otherSpelling string, atrs *languagetool.AnalyzedTokenReadings) string {
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

func (r *WordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
