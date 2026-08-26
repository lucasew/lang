package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDashRule_UpperAfterHyphen(t *testing.T) {
	r := NewDashRule(nil)
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("Diäten-", nil, nil),
		}, 0),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("Erhöhung", nil, nil),
		}, 7),
	}
	ms := r.Match(languagetool.NewAnalyzedSentence(toks))
	require.Len(t, ms, 1)
	// lower after hyphen → no match
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("Nord-", nil, nil),
		}, 0),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("süd", nil, nil),
		}, 5),
	}
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence(toks2)))
	// UND exception
	toks3 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("NORD-", nil, nil),
		}, 0),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("UND", nil, nil),
		}, 5),
	}
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence(toks3)))
}

func TestCompoundCoherencyLemma_HyphenUTF16(t *testing.T) {
	lemma := "Jugendfoto"
	surface := "Jugend-Fotos"
	pos := "SUB:NOM:SIN:NEU"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(surface, &pos, &lemma),
	}, 0)
	got := compoundCoherencyLemma(atr)
	require.Equal(t, "Jugend-Foto", got)
}

func TestLineExpander_LowercaseFormGate(t *testing.T) {
	e := NewLineExpander()
	e.VerbForms = func(string) []string { return []string{"machen", "ßalt", "Mach"} }
	got := e.ExpandLine("rüber_machen")
	require.Contains(t, got, "rübermachen")
	require.NotContains(t, got, "rüberßalt")
	require.NotContains(t, got, "rüberMach")
}
