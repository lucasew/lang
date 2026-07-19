package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrFSE(s string) *string { return &s }

func atrFSE(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrFSE(pos), ptrFSE(lemma)), start)
}

func sentenceFSE(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrFSE(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestFindSuggestionsEsFilter_Rewrite(t *testing.T) {
	f := NewFindSuggestionsEsFilter()
	got := f.RewriteEsSuggestions([]struct{ Form, POS string }{
		{"casa", "NCFS000"},
		{"canta", "VMIP3S0"},
		{"obre", "VMIP3S0"}, // vowel → skip es+
		{"foo", "XXXX"},
	}, 10)
	require.Equal(t, []string{"és casa", "es canta"}, got)
}

func TestFindSuggestionsEsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.FindSuggestionsEsFilter"))
}

func TestFindSuggestionsEsFilter_Accept(t *testing.T) {
	f := NewFindSuggestionsEsFilter()
	f.SpellingSuggestions = func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"casa", "canta", "obre"}
	}
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		switch word {
		case "casa":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSE("NCFS000"), ptrFSE("casa")))
		case "canta":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSE("VMIP3S00"), ptrFSE("cantar")))
		case "obre":
			return languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(word, ptrFSE("VMIP3S00"), ptrFSE("obrir")))
		default:
			return nil
		}
	}
	// es csa → misspelling
	es := atrFSE("es", "PP3CN000", "es", 0)
	csa := atrFSE("csa", "UNKNOWN", "csa", 3)
	csa.SetWhitespaceBefore(true)
	sent := sentenceFSE(es, csa)
	m := rules.NewRuleMatch(nil, sent, 0, csa.GetEndPos(), "msg")
	// pattern tokens like sentence non-blank without SENT for scan - use full non-blank
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Contains(t, sugs, "és casa")
	require.Contains(t, sugs, "es canta")
	// obre is vowel-initial → no "es obre"
	for _, s := range sugs {
		require.NotContains(t, s, "es obre")
	}
	require.Contains(t, out.GetMessage(), "És")
	require.Contains(t, out.GetMessage(), "Es")
}

func TestFindSuggestionsEsFilter_EsAccentOnly(t *testing.T) {
	// és + only nominal suggestions → drop match (spelling rule handles it)
	f := NewFindSuggestionsEsFilter()
	f.SpellingSuggestions = func(atr *languagetool.AnalyzedTokenReadings) []string {
		return []string{"casa"}
	}
	f.Tag = func(word string) *languagetool.AnalyzedTokenReadings {
		return languagetool.NewAnalyzedTokenReadings(
			languagetool.NewAnalyzedToken(word, ptrFSE("NCFS000"), ptrFSE("casa")))
	}
	es := atrFSE("és", "VSIP3S00", "ser", 0)
	csa := atrFSE("csa", "UNKNOWN", "csa", 3)
	csa.SetWhitespaceBefore(true)
	sent := sentenceFSE(es, csa)
	m := rules.NewRuleMatch(nil, sent, 0, csa.GetEndPos(), "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}
