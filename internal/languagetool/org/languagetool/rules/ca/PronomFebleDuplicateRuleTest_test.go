package ca

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/rules/ca/PronomFebleDuplicateRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of PronomFebleDuplicateRuleTest.testRule — simplified POS-driven cases.
func TestPronomFebleDuplicateRule_Rule(t *testing.T) {
	r := NewPronomFebleDuplicateRule(nil)
	require.Equal(t, "PRONOMS_FEBLES_DUPLICATS", r.GetID())
	require.Equal(t, "Pronoms febles duplicats", r.GetDescription())

	// em vaig-me → pronoun before + clitic after conjugated verb
	ss := languagetool.SentenceStartTagName
	vtag := "VMIP1S0"
	sent := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("em", nil, 0, true),
		tok("vaig", &vtag, 3, true),
		tok("-me", nil, 7, false),
	)
	matches := r.Match(sent)
	require.Len(t, matches, 1)
	require.NotEmpty(t, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "pronoms febles")

	// only before: em vaig → no match
	sent2 := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("em", nil, 0, true),
		tok("vaig", &vtag, 3, true),
	)
	require.Empty(t, r.Match(sent2))

	// only after: vaig-me → no match (no before)
	sent3 := buildPronomSentence(
		tok("", &ss, 0, false),
		tok("vaig", &vtag, 0, true),
		tok("-me", nil, 4, false),
	)
	require.Empty(t, r.Match(sent3))
}

func TestIsPronomFebleToken(t *testing.T) {
	require.True(t, isPronomFebleToken(tok("em", nil, 0, true)))
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
