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

	// Surface RE + word POS without tagger (TL ADJECTIVE-V_COMMON_NOUN)
	adj := NewPatternToken(".*[aeiou]", false, true, false)
	adj.Pos = &PosToken{PosTag: "(ADMO|ADCO).*", Regexp: true}
	am := NewPatternTokenMatcher(adj)
	require.True(t, am.IsMatched(languagetool.NewAnalyzedToken("mababa", nil, nil)))
	require.False(t, am.IsMatched(languagetool.NewAnalyzedToken("madasalin", nil, nil)))
}

func TestSoftRegexpAlternatives(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, softRegexpAlternatives("a|b|c"))
	require.Equal(t, []string{"foo", "bar"}, softRegexpAlternatives("(?:foo|bar)"))
}

func TestSoftIrregularLemma(t *testing.T) {
	require.True(t, softInflectedSurfaceMatch("was", "be", false))
	require.True(t, softInflectedSurfaceMatch("est", "être", false))
	require.True(t, softInflectedSurfaceMatch("va", "dir", false)) // AST
	require.True(t, softInflectedSurfaceMatch("va", "ir", false))  // ES
	require.True(t, softInflectedSurfaceMatch("va", "aller", false))
	require.True(t, softInflectedSurfaceMatch("ist", "sein", false))
	require.False(t, softInflectedSurfaceMatch("va", "be", false))
}

func TestSoftGermanGeParticiple(t *testing.T) {
	require.True(t, softGermanGeParticiple("gemacht", "machen"))
	require.True(t, softGermanGeParticiple("gelernt", "lernen"))
	require.True(t, softInflectedSurfaceMatch("gemacht", "machen", false))
	require.True(t, softInflectedSurfaceMatch("genommen", "nehmen", false)) // irregular map
	require.True(t, softInflectedSurfaceMatch("ging", "gehen", false) || softGermanGeParticiple("gegangen", "gehen"))
	require.True(t, softInflectedSurfaceMatch("gegangen", "gehen", false) || softGermanGeParticiple("gegangen", "gehen"))
}

func TestSoftClosedClassPOS(t *testing.T) {
	// DT_PRP: empty PRP$ must not soft-match nouns.
	prp := Pos("PRP$?")
	prp.Pos.Regexp = true
	m := NewPatternTokenMatcher(prp)
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("my", nil, nil)))
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("you", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("search", nil, nil)))

	// Open-class empty POS still soft-matches words (NN-like).
	nn := Pos("NN.*")
	nn.Pos.Regexp = true
	nm := NewPatternTokenMatcher(nn)
	require.True(t, nm.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))
}

// Upstream EN NON_ENGLISH_CHARACTER_IN_A_WORD uses Java \uXXXX escapes.
func TestSoftNormalizeJavaRegexpUnicode(t *testing.T) {
	pat := `[a-z]*(\u043E|\u0455|\u0435|\u0440|\u03BF)[a-z]*`
	got := softNormalizeJavaRegexp(pat)
	require.Contains(t, got, `\x{043e}`)
	require.Contains(t, got, `\x{0455}`)
	m := NewPatternTokenMatcher(NewPatternToken(pat, false, true, false))
	// U+0455 CYRILLIC SMALL LETTER DZE looks like Latin s
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("ѕee", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("see", nil, nil)))
}
