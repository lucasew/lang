package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrDT(s string) *string { return &s }

func atrDT(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrDT(pos), ptrDT(lemma)), start)
}

func sentenceDT(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrDT(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestDonarTempsSuggestionsFilter_Suggest(t *testing.T) {
	f := NewDonarTempsSuggestionsFilter()
	f.SynthHaver = func(suffix string) string { return "ha" }
	f.SynthTenir = func(postag string) string { return "tinc" }
	got := f.Suggest(DonarTempsInput{
		PronomGenderNumber: "1S",
		VerbPostag:         "VMIP3S00",
		CasingModel:        "em",
	})
	require.Contains(t, got, "hi ha temps")
	require.Contains(t, got, "tinc temps")
}

func TestPronomGenderNumberFromP(t *testing.T) {
	require.Equal(t, "1S", PronomGenderNumberFromP("PP1CS000"))
}

func TestDonarTempsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.DonarTempsSuggestionsFilter"))
}

// "em dóna temps" → "hi ha temps", "tinc temps"
// non-blank: [0]SENT [1]em [2]dóna [3]temps
func TestDonarTempsSuggestionsFilter_AcceptDirect(t *testing.T) {
	f := NewDonarTempsSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		switch {
		case lem == "haver" && postag == "VAIP3S00":
			return []string{"ha"}
		case lem == "tenir" && postag == "VMIP1S00":
			return []string{"tinc"}
		default:
			return nil
		}
	}
	// dóna: VMIP3S00 → VA + IP3S00 = VAIP3S00; tenir: VMIP + 1S + 00 = VMIP1S00
	sent := sentenceDT(
		atrDT("em", "PP1CS000", "em", 0),
		atrDT("dóna", "VMIP3S00", "donar", 3),
		atrDT("temps", "NCMS000", "temps", 8),
	)
	m := rules.NewRuleMatch(nil, sent, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Len(t, sugs, 2)
	require.Equal(t, "hi ha temps", sugs[0])
	require.Equal(t, "tinc temps", sugs[1])
	require.Equal(t, 0, out.GetFromPos())
	require.Equal(t, sent.GetTokensWithoutWhitespace()[3].GetEndPos(), out.GetToPos())
}

func TestDonarTempsSuggestionsFilter_NoSynth(t *testing.T) {
	f := NewDonarTempsSuggestionsFilter()
	sent := sentenceDT(
		atrDT("em", "PP1CS000", "em", 0),
		atrDT("dóna", "VMIP3S00", "donar", 3),
		atrDT("temps", "NCMS000", "temps", 8),
	)
	m := rules.NewRuleMatch(nil, sent, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}

// "em va donar temps" with aux — re-inflect va to 1S + tenir
func TestDonarTempsSuggestionsFilter_AcceptWithAux(t *testing.T) {
	f := NewDonarTempsSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		// va: VAIP3S00 → VAIP1S00 = vaig; donar VMN0000 → tenir; haver VA + MN0000?
		// verbPostag of donar if VMN00000 or similar
		switch {
		case lem == "haver":
			return []string{"haver"} // VA + from donar postag
		case lem == "tenir":
			return []string{"donat"} // exact main postag synth for tenir
		case postag == "VAIP1S00":
			return []string{"vaig"}
		default:
			return nil
		}
	}
	// tokens: em va donar temps
	// va: VAIP3S00 lemma anar? actually "va" is VAIP3S00 of anar/auxiliar
	sent := sentenceDT(
		atrDT("em", "PP1CS000", "em", 0),
		atrDT("va", "VAIP3S00", "anar", 3),
		atrDT("donar", "VMN00000", "donar", 6),
		atrDT("temps", "NCMS000", "temps", 12),
	)
	// whitespace before aux middles
	for i, tok := range sent.GetTokensWithoutWhitespace() {
		if i >= 2 {
			tok.SetWhitespaceBefore(true)
		}
	}
	m := rules.NewRuleMatch(nil, sent, 0, 11, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	// at least haver-hi path
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs[0], "hi")
	require.Contains(t, sugs[0], "temps")
}
