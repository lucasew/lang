package de

// Twin of GermanDisambiguationTest — chunker + ignore-spelling green slices
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	disde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/de"
	"github.com/stretchr/testify/require"
)

func atrDE(token, pos string) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(token, &p, nil)}, 0)
}

// Port of GermanDisambiguationTest.testChunker
func TestGermanDisambiguation_Chunker(t *testing.T) {
	toks2 := []*languagetool.AnalyzedTokenReadings{
		atrDE("ein", "ART:IND:AKK:SIN:NEU"),
		atrDE("Konzept", "SUB:AKK:SIN:NEU"),
	}
	ch := chunking.NewGermanChunker()
	ch.AddChunkTags(toks2)
	require.Equal(t, []string{"B-NP"}, toks2[0].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, toks2[1].GetChunkTags())

	// "3-adische System" — digit-hyphen adj ignore-by-speller
	toks3 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("3-adische", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("System", nil, nil)),
	}
	d := disde.NewGermanRuleDisambiguator()
	// Test-only stage: digit-hyphen ignore pattern (not production invent).
	d.MultitokenIgnore = ignoreSpellingStep{}
	sent := languagetool.NewAnalyzedSentence(toks3)
	out := d.Disambiguate(sent)
	require.True(t, out.GetTokens()[0].IsIgnoredBySpeller())
	// "System" not matched by digit-hyphen pattern alone
	require.False(t, out.GetTokens()[2].IsIgnoredBySpeller())

	// Kelassurier multiword list soft: without dict, not ignored
	toks4 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Kelassurier", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Mauer", nil, nil)),
	}
	out4 := d.Disambiguate(languagetool.NewAnalyzedSentence(toks4))
	require.False(t, out4.GetTokens()[0].IsIgnoredBySpeller(), "Kelassurier needs multitoken dict (soft)")
}
