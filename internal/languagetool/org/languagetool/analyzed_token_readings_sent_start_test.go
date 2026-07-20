package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Disambiguation REPLACE must keep the sentence-start slot (Java nonBlank[0]=SENT_START).
func TestFromOld_PreservesSentenceStart(t *testing.T) {
	ss := SentenceStartTagName
	start := NewAnalyzedTokenReadings(NewAnalyzedToken("", &ss, nil))
	require.True(t, start.IsSentenceStart())
	require.True(t, start.IsWhitespace())

	z := "Z0MP0"
	repl := NewAnalyzedToken("", &z, nil)
	out := NewAnalyzedTokenReadingsFromOld(start, []*AnalyzedToken{repl}, "RULE")
	require.True(t, out.IsSentenceStart(), "REPLACE must not drop isSentenceStart on SENT_START slot")
	// Still kept in non-blank list
	s := NewAnalyzedSentence([]*AnalyzedTokenReadings{
		out,
		NewAnalyzedTokenReadingsAt(NewAnalyzedToken("teste", nil, nil), 0),
		NewAnalyzedTokenReadingsAt(NewAnalyzedToken("teste", nil, nil), 6),
	})
	nws := s.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(nws), 3)
	require.True(t, nws[0].IsSentenceStart())
}
