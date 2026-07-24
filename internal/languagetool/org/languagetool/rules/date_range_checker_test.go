package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateRangeChecker(t *testing.T) {
	c := NewDateRangeChecker()
	require.False(t, c.Accept("1990", "2000")) // valid range → suppress
	require.True(t, c.Accept("2000", "1990"))  // invalid → keep
	require.False(t, c.Accept("ab", "2000"))
}

func TestShortenedYearRangeChecker(t *testing.T) {
	c := NewShortenedYearRangeChecker()
	// 1998-99 → 1999, 1998 < 1999 → valid → suppress
	require.False(t, c.Accept("1998", "99"))
	// 1998-92 → 1992, 1998 >= 1992 → invalid → keep
	require.True(t, c.Accept("1998", "92"))
	require.False(t, c.Accept("x", "92"))
	m := NewRuleMatch(nil, nil, 0, 5, "msg")
	require.NotNil(t, c.AcceptRuleMatch(m, map[string]string{"x": "1998", "y": "92"}, 0, nil, nil))
	require.Nil(t, c.AcceptRuleMatch(m, map[string]string{"x": "1998", "y": "99"}, 0, nil, nil))
}

func TestMatchPositionAndSuggestion(t *testing.T) {
	p := NewMatchPosition(1, 5)
	require.Equal(t, "1-5", p.String())
	s := SuggestionWithMessage{Suggestion: "fix", Message: "msg"}
	require.Equal(t, "fix", s.Suggestion)
	require.Equal(t, "msg", s.Message)
}
