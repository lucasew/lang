package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternTokenMatcher(t *testing.T) {
	m := NewPatternTokenMatcher(Token("Hello"))
	tok := languagetool.NewAnalyzedToken("hello", nil, nil)
	require.True(t, m.IsMatched(tok))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("world", nil, nil)))

	cs := NewPatternTokenMatcher(CsToken("Hello"))
	require.False(t, cs.IsMatched(tok))
	require.True(t, cs.IsMatched(languagetool.NewAnalyzedToken("Hello", nil, nil)))

	re := NewPatternTokenMatcher(TokenRegex("c.t"))
	require.True(t, re.IsMatched(languagetool.NewAnalyzedToken("cat", nil, nil)))

	pos := "NN"
	pm := NewPatternTokenMatcher(Pos("NN"))
	require.True(t, pm.IsMatched(languagetool.NewAnalyzedToken("dog", &pos, nil)))
}
