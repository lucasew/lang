package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrOB(s string) *string { return &s }

func atrOB(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrOB(pos), ptrOB(lemma)), start)
}

func sentenceOB(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrOB(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestOblidarseSugestionsFilter_Prefix(t *testing.T) {
	f := NewOblidarseSugestionsFilter()
	require.True(t, f.NeedsApostrophe("oblidat"))
	require.False(t, f.NeedsApostrophe("passat"))
	require.Equal(t, "m'", f.ReflexivePrefix("1S", true, false))
	require.Equal(t, "em ", f.ReflexivePrefix("1S", false, false))
	require.Equal(t, "me n'", f.ReflexivePrefix("1S", true, true))
	require.Equal(t, "me'n ", f.ReflexivePrefix("1S", false, true))
}

func TestOblidarseRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.OblidarseSugestionsFilter"))
}

// se m'ha oblidat → en form (me n'he oblidat)
func TestOblidarseSugestionsFilter_AcceptEnForm(t *testing.T) {
	f := NewOblidarseSugestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		if lem == "haver" && postag == "VAIP1S00" {
			return []string{"he"}
		}
		return nil
	}

	se := atrOB("se", "P00CN000", "es", 0)
	mp := atrOB("m'", "PP1CS000", "jo", 3)
	mp.SetWhitespaceBeforeToken(" ")
	ha := atrOB("ha", "VAIP3S00", "haver", 6)
	ha.SetWhitespaceBeforeToken(" ")
	ob := atrOB("oblidat", "VMP00SM", "oblidar", 9)
	ob.SetWhitespaceBeforeToken(" ")

	sent := sentenceOB(se, mp, ha, ob)
	match := rules.NewRuleMatch(nil, sent, 0, 8, "msg")
	out := f.AcceptRuleMatch(match, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs[0], "he")
	require.Contains(t, sugs[0], "oblidat")
	// en form prefix
	require.True(t, strings.Contains(sugs[0], "n'") || strings.Contains(sugs[0], "me n") || strings.Contains(sugs[0], "'n"),
		"expected en-form prefix, got %q", sugs[0])
}

// se m'ha oblidat de → without en
func TestOblidarseSugestionsFilter_AcceptWithDe(t *testing.T) {
	f := NewOblidarseSugestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		if postag == "VAIP1S00" {
			return []string{"he"}
		}
		return nil
	}
	se := atrOB("se", "P00CN000", "es", 0)
	mp := atrOB("m'", "PP1CS000", "jo", 3)
	mp.SetWhitespaceBeforeToken(" ")
	ha := atrOB("ha", "VAIP3S00", "haver", 6)
	ha.SetWhitespaceBeforeToken(" ")
	ob := atrOB("oblidat", "VMP00SM", "oblidar", 9)
	ob.SetWhitespaceBeforeToken(" ")
	de := atrOB("de", "SPS00", "de", 17)
	de.SetWhitespaceBeforeToken(" ")

	sent := sentenceOB(se, mp, ha, ob, de)
	match := rules.NewRuleMatch(nil, sent, 0, 8, "msg")
	out := f.AcceptRuleMatch(match, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs[0], "he")
	require.NotContains(t, sugs[0], "n'")
	require.NotContains(t, sugs[0], "me n")
}

func TestOblidarseSugestionsFilter_NoSynth(t *testing.T) {
	f := NewOblidarseSugestionsFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
	sent := sentenceOB(atrOB("se", "P00CN000", "es", 0))
	m := rules.NewRuleMatch(nil, sent, 0, 2, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}
