package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrDB(s string) *string { return &s }

func atrDB(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrDB(pos), ptrDB(lemma)), start)
}

func sentenceDB(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrDB(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestDonarseliBeFilter_Helpers(t *testing.T) {
	require.Equal(t, "malament", NormalizeAdverbi("mal"))
	require.True(t, IsAdverbiFinal("bé"))
	require.True(t, IsPronomPersonal("mi"))
	require.True(t, IsExceptionQue("ja"))
	f := NewDonarseliBeFilter()
	s := f.BuildDonarSuggestion("se", "dona", true, "")
	require.Contains(t, s, "dona")
}

func TestDonarseliBeRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.DonarseliBeFilter"))
}

// se li dona bé — two pronouns before (se, li), verb dona, adverb bé
// tokens: [0]SENT [1]se [2]li [3]dona [4]bé
// posInit at se (index 1); posPrimerVerb=3; numPronounsBefore=2; relevant=posPrimerVerb-(2-1)=2 → li
func TestDonarseliBeFilter_AcceptBasic(t *testing.T) {
	f := NewDonarseliBeFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		// newVerbPostag from dona VMIP3S00 + li person3 number S → still 3S-ish
		// li is PP3CSD00: person 3, number C at [4]? PP3CSD00: [2]=3 [4]=S
		switch lem {
		case "tenir":
			return []string{"té"}
		case "fer":
			return []string{"fa"}
		case "sortir":
			return []string{"surto"}
		case "anar":
			return []string{"va"}
		case "eixir":
			return []string{"ix"}
		default:
			return []string{"X"}
		}
	}

	// Weak pronouns matching pPronomFeble
	// se: P00CN000 or similar P0.{6}
	// li: PP3CSD00
	se := atrDB("se", "P00CN000", "es", 0)
	li := atrDB("li", "PP3CSD00", "ell", 3)
	li.SetWhitespaceBefore(true)
	dona := atrDB("dona", "VMIP3S00", "donar", 6)
	dona.SetWhitespaceBefore(true)
	be := atrDB("bé", "RG", "bé", 11)
	be.SetWhitespaceBefore(true)

	sent := sentenceDB(se, li, dona, be)
	// match span se..dona end
	m := rules.NewRuleMatch(nil, sent, 0, dona.GetEndPos(), "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// expect several paraphrases containing traça / bé / etc.
	joined := strings.Join(sugs, " | ")
	require.Contains(t, joined, "traça")
	require.Contains(t, joined, "bé")
}

func TestDonarseliBeFilter_NoSynth(t *testing.T) {
	f := NewDonarseliBeFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}

func TestDonarseliBeFilter_NoAdverb(t *testing.T) {
	f := NewDonarseliBeFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		return []string{"x"}
	}
	se := atrDB("se", "P00CN000", "es", 0)
	li := atrDB("li", "PP3CSD00", "ell", 3)
	li.SetWhitespaceBefore(true)
	dona := atrDB("dona", "VMIP3S00", "donar", 6)
	dona.SetWhitespaceBefore(true)
	// no final adverb
	sent := sentenceDB(se, li, dona)
	m := rules.NewRuleMatch(nil, sent, 0, dona.GetEndPos(), "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}

func TestGetAdverbsFor_Traca(t *testing.T) {
	molt := atrDB("molt", "RG", "molt", 0)
	molt.SetWhitespaceBefore(true)
	// primer=0 darrer=1 with token "molt" → " molt" if whitespace before
	got := getAdverbsFor([]*languagetool.AnalyzedTokenReadings{molt}, 0, 1, "traça")
	require.Equal(t, " molta", got)
}
