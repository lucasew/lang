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

// Faithful: untagged tokens do not soft-accept open-class POS patterns.
func TestPatternTokenMatcher_UntaggedPOSStrict(t *testing.T) {
	nn := Pos("NN.*")
	nn.Pos.Regexp = true
	nm := NewPatternTokenMatcher(nn)
	require.False(t, nm.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))

	unk := Pos("UNKNOWN")
	um := NewPatternTokenMatcher(unk)
	require.True(t, um.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))
}

// Upstream EN NON_ENGLISH_CHARACTER_IN_A_WORD uses Java \uXXXX escapes.
func TestNormalizeJavaRegexpUnicode(t *testing.T) {
	pat := `[a-z]*(\u043E|\u0455|\u0435|\u0440|\u03BF)[a-z]*`
	got := normalizeJavaRegexp(pat)
	require.Contains(t, got, `\x{043e}`)
	require.Contains(t, got, `\x{0455}`)
	m := NewPatternTokenMatcher(NewPatternToken(pat, false, true, false))
	// U+0455 CYRILLIC SMALL LETTER DZE looks like Latin s
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("ѕee", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("see", nil, nil)))
}

func TestIsMatchedReadings_ChunkTag(t *testing.T) {
	pt := NewPatternToken("house", false, false, false)
	pt.SetChunkTag("B-NP", false)
	m := NewPatternTokenMatcher(pt)
	pos := "NN"
	atr := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("house", &pos, nil)},
		0,
	)
	atr.SetChunkTags([]string{"B-NP"})
	require.True(t, m.IsMatchedReadings(atr))
	atr.SetChunkTags([]string{"I-VP"})
	require.False(t, m.IsMatchedReadings(atr))
}
