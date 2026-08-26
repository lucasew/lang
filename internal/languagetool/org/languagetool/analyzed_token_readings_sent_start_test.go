package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of Java AnalyzedTokenReadings(old, newReadings, ruleApplied):
// isSentStart comes only from new readings[0] POS, not from old.
func TestFromOld_SentenceStartFromNewReadingsOnly(t *testing.T) {
	ss := SentenceStartTagName
	start := NewAnalyzedTokenReadings(NewAnalyzedToken("", &ss, nil))
	require.True(t, start.IsSentenceStart())
	require.True(t, start.IsWhitespace())

	// REPLACE without SENT_START POS → Java drops isSentenceStart
	z := "Z0MP0"
	repl := NewAnalyzedToken("", &z, nil)
	out := NewAnalyzedTokenReadingsFromOld(start, []*AnalyzedToken{repl}, "RULE")
	require.False(t, out.IsSentenceStart(), "Java FromOld does not preserve isSentStart without SENT_START POS")

	// REPLACE that keeps SENT_START as reading[0] POS → still sentence start
	keep := NewAnalyzedTokenReadingsFromOld(start, []*AnalyzedToken{NewAnalyzedToken("", &ss, nil)}, "RULE")
	require.True(t, keep.IsSentenceStart())

	// empty surface + isSentStart false is whitespace-only → dropped from nonBlank
	// unless isSentenceStart/End/ParaEnd (Java getTokensWithoutWhitespace).
	s := NewAnalyzedSentence([]*AnalyzedTokenReadings{
		out,
		NewAnalyzedTokenReadingsAt(NewAnalyzedToken("teste", nil, nil), 0),
		NewAnalyzedTokenReadingsAt(NewAnalyzedToken("teste", nil, nil), 6),
	})
	nws := s.GetTokensWithoutWhitespace()
	// empty untagged slot is whitespace → not kept; no soft invent preserve
	require.False(t, len(nws) > 0 && nws[0].IsSentenceStart())
}
