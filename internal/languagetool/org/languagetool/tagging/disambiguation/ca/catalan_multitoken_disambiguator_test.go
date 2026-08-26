package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanMultitokenDisambiguator(t *testing.T) {
	d := NewCatalanMultitokenDisambiguator()
	// no speller → identity
	s := languagetool.AnalyzePlain("Hola món")
	require.Equal(t, s.GetText(), d.Disambiguate(s).GetText())

	d.IsMisspelled = func(phrase string) bool {
		return phrase != "Santa Maria"
	}
	// untagged tokens with phrase accepted by speller
	ss := languagetool.SentenceStartTagName
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Santa", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Maria", nil, nil)),
	}
	// mark space as whitespace-like by token content — IsWhitespace uses tools.IsWhitespace
	sent := languagetool.NewAnalyzedSentence(tokens)
	out := d.Disambiguate(sent)
	require.NotNil(t, out)
	// Santa or Maria should gain NPCNM00 if phrase matched
	readings := out.GetTokens()[1].GetReadings()
	found := false
	for _, r := range readings {
		if r != nil && r.GetPOSTag() != nil && *r.GetPOSTag() == "NPCNM00" {
			found = true
		}
	}
	require.True(t, found)
}
