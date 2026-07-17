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

// Soft path: upstream goldens only (RU kurica, RU software adj, TA polum).
func TestSoftInflectedAndSurfacePOS(t *testing.T) {
	// яйцо / яйца — short shared stem (min 3)
	require.True(t, softSharedStemMatch("яйца", "яйцо"))
	require.True(t, softInflectedSurfaceMatch("высиживает", "высиживать", false))

	// RE alternatives with inflected="yes" (Adj_NN_number_Software)
	pt := NewPatternToken("программный|аппаратный", false, true, true)
	m := NewPatternTokenMatcher(pt)
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("программных", nil, nil)))

	// Surface "." with postag SENT_END (TA polum second token)
	sentEnd := NewPatternToken(".", false, false, false)
	sentEnd.Pos = &PosToken{PosTag: "SENT_END"}
	sm := NewPatternTokenMatcher(sentEnd)
	require.True(t, sm.IsMatched(languagetool.NewAnalyzedToken(".", nil, nil)))
}

func TestSoftRegexpAlternatives(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, softRegexpAlternatives("a|b|c"))
	require.Equal(t, []string{"foo", "bar"}, softRegexpAlternatives("(?:foo|bar)"))
}
