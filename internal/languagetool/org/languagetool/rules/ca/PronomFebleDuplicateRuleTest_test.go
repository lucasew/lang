package ca

// Twin of PronomFebleDuplicateRuleTest — POS-only pronouns (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPronomFebleDuplicateRule_Rule(t *testing.T) {
	r := NewPronomFebleDuplicateRule(nil)
	require.Equal(t, "PRONOMS_FEBLES_DUPLICATS", r.GetID())
	require.Equal(t, "Pronoms febles duplicats", r.GetDescription())

	// em vaig-me → pronoun before + clitic after conjugated verb
	ss := languagetool.SentenceStartTagName
	vtag := "VMIP1S0"
	p0 := "P01CN000" // matches P0.{6}
	// enclitic -me: PP3..A00 style — use PP3CSA00 or P0
	pMe := "P00CN000"

	sent := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("em", &p0, 0, true),
		tok("vaig", &vtag, 3, true),
		tok("-me", &pMe, 7, false),
	)
	matches := r.Match(sent)
	require.Len(t, matches, 1)
	require.NotEmpty(t, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "pronoms febles")

	// only before: em vaig → no match
	sent2 := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("em", &p0, 0, true),
		tok("vaig", &vtag, 3, true),
	)
	require.Empty(t, r.Match(sent2))

	// only after: vaig-me → no match (no before)
	sent3 := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("vaig", &vtag, 0, true),
		tok("-me", &pMe, 4, false),
	)
	require.Empty(t, r.Match(sent3))

	// without pronoun POS: fail closed
	sent4 := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("em", nil, 0, true),
		tok("vaig", &vtag, 3, true),
		tok("-me", nil, 7, false),
	)
	require.Empty(t, r.Match(sent4))
}

func TestIsPronomFebleToken(t *testing.T) {
	require.False(t, isPronomFebleToken(tok("em", nil, 0, true)), "no surface invent")
	tag := "P0000000"
	require.True(t, isPronomFebleToken(tok("x", &tag, 0, true)))
	require.False(t, isPronomFebleToken(tok("casa", nil, 0, true)))
}

func tok(surface string, pos *string, start int, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	at := languagetool.NewAnalyzedToken(surface, pos, nil)
	at.SetWhitespaceBefore(wsBefore)
	r := languagetool.NewAnalyzedTokenReadingsAt(at, start)
	return r
}

func buildPronomSentence(tokens ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(tokens)
}
