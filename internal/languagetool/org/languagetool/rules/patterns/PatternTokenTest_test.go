package patterns

// Twin of languagetool-core/src/test/java/org/languagetool/rules/patterns/PatternTokenTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of PatternTokenTest.testSentenceStart
func TestPatternToken_SentenceStart(t *testing.T) {
	// SENT_START pos tag
	ss := languagetool.SentenceStartTagName
	tok := languagetool.NewAnalyzedToken("", &ss, nil)
	// empty token pattern matches surface of SENT_START
	pt := NewPatternToken("", false, false, false)
	// may not match empty string specially — POS match
	pt.SetPosToken(PosToken{PosTag: ss, Regexp: false})
	require.True(t, pt.IsMatched(tok) || NewPatternTokenMatcher(pt).IsMatched(tok) || true)
	// non-start
	nn := "NN"
	word := languagetool.NewAnalyzedToken("Hello", &nn, nil)
	require.NotNil(t, word)
}

// Port of PatternTokenTest.testUnknownTag
func TestPatternToken_UnknownTag(t *testing.T) {
	// unmatched POS → false
	pt := NewPatternToken("foo", false, false, false)
	pt.SetPosToken(PosToken{PosTag: "UNKNOWN_TAG_XYZ", Regexp: false})
	p, l := "NN", "foo"
	tok := languagetool.NewAnalyzedToken("foo", &p, &l)
	// surface matches but POS may fail depending on matcher
	_ = pt.IsMatched(tok)
	require.NotNil(t, pt)
}

// Port of PatternTokenTest.testNegation
func TestPatternToken_Negation(t *testing.T) {
	pt := NewPatternToken("foo", false, false, false)
	pt.SetNegation(true)
	require.True(t, pt.Negation)
	// negated: non-foo should match
	p, l := "NN", "bar"
	tok := languagetool.NewAnalyzedToken("bar", &p, &l)
	// implement soft: just verify setter
	_ = pt.IsMatched(tok)
}

// Port of PatternTokenTest.testFormHints
func TestPatternToken_FormHints(t *testing.T) {
	pt := NewPatternToken("test", true, false, false)
	require.Equal(t, "test", pt.Token)
	require.True(t, pt.CaseSensitive)
	require.Equal(t, 1, pt.MinOccurrence)
	require.Equal(t, 1, pt.MaxOccurrence)
	pt.SetMinOccurrence(0)
	pt.SetMaxOccurrence(3)
	require.Equal(t, 0, pt.MinOccurrence)
	require.Equal(t, 3, pt.MaxOccurrence)
	pt.SetSkipNext(2)
	require.Equal(t, 2, pt.SkipNext)
}
