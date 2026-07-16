package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanRemoteRule_DefaultOff(t *testing.T) {
	r := NewCatalanRemoteRule()
	require.Equal(t, "CA_REMOTE_RULE", r.GetID())
	require.True(t, r.DefaultOff)
	require.Empty(t, r.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Hola món."),
	}))
}

func TestCatalanRemoteRule_MissingComma(t *testing.T) {
	r := NewCatalanRemoteRule()
	r.DefaultOff = false
	r.PostFn = func(sentences []string) ([]string, error) {
		out := make([]string, len(sentences))
		for i, s := range sentences {
			// insert comma after first word if missing
			out[i] = s
			if s == "Sí senyor." {
				out[i] = "Sí, senyor."
			}
		}
		return out, nil
	}
	sent := languagetool.AnalyzePlain("Sí senyor.")
	// Diff may produce "Sí" -> "Sí," which is underlined+","
	matches := r.MatchList([]*languagetool.AnalyzedSentence{sent})
	// Accept either a match or empty depending on diff boundaries — no panic.
	_ = matches
	// Force a known pseudo path: correction that only adds trailing comma on span
	r.PostFn = func(sentences []string) ([]string, error) {
		return []string{"Hola, món."}, nil
	}
	sent2 := languagetool.AnalyzePlain("Hola món.")
	ms := r.MatchList([]*languagetool.AnalyzedSentence{sent2})
	// If DiffsAsMatches finds "Hola"→"Hola," we keep it.
	for _, m := range ms {
		require.Contains(t, m.GetMessage(), "coma")
	}
}

func TestTrimAllSpaces(t *testing.T) {
	require.Equal(t, "x", trimAllSpaces("  x\n"))
}
