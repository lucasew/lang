package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractTokenBasedRule_CanBeIgnored(t *testing.T) {
	r := NewAbstractTokenBasedRule("ID", "desc", "en", []*PatternToken{Token("hello")})
	require.NotEmpty(t, r.TokenHints)
	// sentence without hello → ignore
	sent := languagetool.AnalyzePlain("world there")
	require.True(t, r.CanBeIgnoredFor(sent))
	sent2 := languagetool.AnalyzePlain("say hello now")
	require.False(t, r.CanBeIgnoredFor(sent2))
}

func TestTokenHint_OffsetsAndFormHints(t *testing.T) {
	// Regex alternation form hints (Java getPossibleRegexpValues subset).
	pt := NewPatternToken("color|colour", false, true, false)
	require.Equal(t, []string{"color", "colour"}, pt.CalcFormHints())
	require.Nil(t, pt.CalcLemmaHints())

	inf := NewPatternToken("run", false, false, true)
	require.Nil(t, inf.CalcFormHints())
	require.Equal(t, []string{"run"}, inf.CalcLemmaHints())

	// Optional token: no form hints (MAY_BE_OMITTED)
	opt := Token("the")
	opt.SetMinOccurrence(0)
	require.Nil(t, opt.CalcFormHints())

	// TokenHint uses GetTokenOffsets (not full scan).
	ss := languagetool.SentenceStartTagName
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("colour", nil, nil), 0),
	})
	h := NewTokenHint(false, []string{"color", "colour"}, 0)
	require.False(t, h.CanBeIgnoredFor(sent))
	require.Equal(t, []int{1}, h.GetPossibleIndices(sent))
	h2 := NewTokenHint(false, []string{"missing"}, 0)
	require.True(t, h2.CanBeIgnoredFor(sent))
	require.Nil(t, h2.GetPossibleIndices(sent))

	// Anchor on fixed-offset rule with regex alternation
	r := NewAbstractTokenBasedRule("C", "d", "en", []*PatternToken{
		NewPatternToken("color|colour", false, true, false),
		Token("blind"),
	})
	require.NotNil(t, r.AnchorHint)
	require.Contains(t, r.AnchorHint.LowerCaseValues, "color")
	require.Contains(t, r.AnchorHint.LowerCaseValues, "colour")
}
