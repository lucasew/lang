package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpaceBeforeMatching(t *testing.T) {
	// spacebefore="no" — must not have whitespace before
	pt := NewPatternToken("s", false, false, false)
	pt.SetWhitespaceBefore(false)
	m := NewPatternTokenMatcher(pt)

	tokWS := languagetool.NewAnalyzedToken("s", nil, nil)
	tokWS.SetWhitespaceBefore(true)
	require.False(t, m.IsMatched(tokWS))

	tokNo := languagetool.NewAnalyzedToken("s", nil, nil)
	tokNo.SetWhitespaceBefore(false)
	require.True(t, m.IsMatched(tokNo))

	// spacebefore="yes"
	pt2 := NewPatternToken("book", false, false, false)
	pt2.SetWhitespaceBefore(true)
	m2 := NewPatternTokenMatcher(pt2)
	require.True(t, m2.IsMatched(tokWSBook()))
	require.False(t, m2.IsMatched(tokNoBook()))
}

func tokWSBook() *languagetool.AnalyzedToken {
	t := languagetool.NewAnalyzedToken("book", nil, nil)
	t.SetWhitespaceBefore(true)
	return t
}
func tokNoBook() *languagetool.AnalyzedToken {
	t := languagetool.NewAnalyzedToken("book", nil, nil)
	t.SetWhitespaceBefore(false)
	return t
}
