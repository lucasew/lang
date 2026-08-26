package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java RuleMatch.setOriginalErrorStr uses String.substring with UTF-16 indices
// (same as AnalyzedTokenReadings start/end). Multi-byte German must not corrupt.
func TestSetOriginalErrorStr_UTF16MultiByte(t *testing.T) {
	s := languagetool.AnalyzePlain("Größe,")
	toks := s.GetTokensWithoutWhitespace()
	// tokens: sentence-start "", "Größe", ","
	require.GreaterOrEqual(t, len(toks), 3)
	groesse := toks[1]
	require.Equal(t, "Größe", groesse.GetToken())
	from, to := groesse.GetStartPos(), groesse.GetEndPos()
	require.Equal(t, 0, from)
	require.Equal(t, 5, to) // 5 UTF-16 units, not 7 UTF-8 bytes

	m := NewRuleMatch(NewFakeRule("X"), s, from, to, "msg")
	m.SetOriginalErrorStr()
	require.Equal(t, "Größe", m.GetOriginalErrorStr())

	// span including comma
	m2 := NewRuleMatch(NewFakeRule("X"), s, from, toks[2].GetEndPos(), "msg")
	m2.SetOriginalErrorStr()
	require.Equal(t, "Größe,", m2.GetOriginalErrorStr())
}

// FromPosSentence preferred over document FromPos (Java).
func TestSetOriginalErrorStr_PrefersSentencePos(t *testing.T) {
	s := languagetool.AnalyzePlain("Größe")
	m := NewRuleMatch(NewFakeRule("X"), s, 99, 104, "msg") // bogus doc pos
	m.FromPosSentence = 0
	m.ToPosSentence = 5
	m.SetOriginalErrorStr()
	require.Equal(t, "Größe", m.GetOriginalErrorStr())
}
