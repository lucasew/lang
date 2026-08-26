package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrSWAD(s string) *string { return &s }

func atrSWAD(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrSWAD(pos), ptrSWAD(lemma)), start)
}

func sentenceSWAD(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrSWAD(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestSynthesizeWithAnyDeterminerFilter_SuggestAll(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	got := f.SuggestAll([]struct{ Form, POS string }{
		{"amic", "NCMS000"},
		{"amiga", "NCFS000"},
	}, "", "MS", "")
	require.Contains(t, got, "l'amic")
	require.Contains(t, got, "l'amiga")
	require.True(t, IsPreposition("de"))
	require.Equal(t, "d", PrepositionKey("de"))
}

func TestSynthesizeWithAnyDeterminerRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.SynthesizeWithAnyDeterminerFilter"))
}

// el amic → suggestions with det+form spanning from el
func TestSynthesizeWithAnyDeterminerFilter_AcceptDA(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	// dummy index 0 content after SENT so firstUnderlinedToken can be 1 for el
	// tokens: [0]SENT [1]el [2]amic
	el := atrSWAD("el", "DA0MS0", "el", 0)
	amic := atrSWAD("amic", "NCMS000", "amic", 3)
	amic.SetWhitespaceBefore(true)
	sent := sentenceSWAD(el, amic)
	// match fromPos on amic
	m := rules.NewRuleMatch(nil, sent, 3, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaSelect": "N.*",
	}, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// firstUnderlinedToken for DA is el at index 1 → uppercase
	require.Contains(t, sugs, "L'amic")
	require.Equal(t, el.GetStartPos(), out.GetFromPos())
	require.Equal(t, 7, out.GetToPos())
}

// with synthesizer expand forms
func TestSynthesizeWithAnyDeterminerFilter_AcceptSynth(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	f.GetPossibleTags = func() []string { return []string{"NCMS000", "NCFS000"} }
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		switch postag {
		case "NCMS000":
			return []string{"amic"}
		case "NCFS000":
			return []string{"amiga"}
		default:
			return nil
		}
	}
	// not sentence-start underline: index of det >= 2
	// [0]SENT [1]veig [2]el [3]amic
	veig := atrSWAD("veig", "VMIP1S00", "veure", 0)
	el := atrSWAD("el", "DA0MS0", "el", 5)
	el.SetWhitespaceBefore(true)
	amic := atrSWAD("amic", "NCMS000", "amic", 8)
	amic.SetWhitespaceBefore(true)
	sent := sentenceSWAD(veig, el, amic)
	m := rules.NewRuleMatch(nil, sent, 8, 12, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaSelect":   "N.*",
		"synthAllForms": "true",
	}, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Contains(t, sugs, "l'amic")
	require.Contains(t, sugs, "l'amiga")
}

// DD determiner (not DA) → synthesize determiner forms via SynthesizeRE
func TestSynthesizeWithAnyDeterminerFilter_AcceptDD(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	f.SynthesizeRE = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		// DD.[CM]S. etc.
		if tok != nil && tok.GetToken() == "aquest" {
			return []string{"aquest", "aquesta"}
		}
		return nil
	}
	veig := atrSWAD("veig", "VMIP1S00", "veure", 0)
	aq := atrSWAD("aquest", "DD0MS0", "aquest", 5)
	aq.SetWhitespaceBefore(true)
	amic := atrSWAD("amic", "NCMS000", "amic", 12)
	amic.SetWhitespaceBefore(true)
	sent := sentenceSWAD(veig, aq, amic)
	m := rules.NewRuleMatch(nil, sent, 12, 16, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"lemmaSelect": "N.*"}, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	// DD path: "aquest amic" style (preserve case of determiner token)
	require.NotEmpty(t, sugs)
	found := false
	for _, s := range sugs {
		if strings.Contains(s, "amic") && strings.Contains(s, "aquest") {
			found = true
			break
		}
	}
	require.True(t, found, "got %v", sugs)
}

func TestSynthesizeWithAnyDeterminerFilter_MissingLemma(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	tok := atrSWAD("x", "VMIP3S00", "x", 0)
	sent := sentenceSWAD(tok)
	m := rules.NewRuleMatch(nil, sent, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{"lemmaSelect": "N.*"}, 0, nil, nil)
	})
}
