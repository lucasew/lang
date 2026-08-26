package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJavaRegexp_NegativeLookahead(t *testing.T) {
	// noun(?!.*alt).* — common UK postag pattern
	m := NewStringMatcher(`noun(?!.*alt).*`, true, true)
	require.NotNil(t, m.javaRE)
	require.True(t, m.Matches("noun:inanim:m:v_naz"))
	require.False(t, m.Matches("noun:inanim:m:v_naz:alt"))
	require.False(t, m.Matches("adj:m:v_naz"))

	// adv(?!p).* — adv but not advp
	m2 := NewStringMatcher(`adv(?!p).*`, true, true)
	require.True(t, m2.Matches("adv"))
	require.True(t, m2.Matches("adv:compb"))
	require.False(t, m2.Matches("advp"))
	require.False(t, m2.Matches("advp:rev"))

	// (?!.*alt).* — whole string without alt
	m3 := NewStringMatcher(`(?!.*alt).*`, true, true)
	require.True(t, m3.Matches("noun:foo"))
	require.False(t, m3.Matches("noun:alt:bar"))

	// adj:.:v_zna(?!:rinanim).*
	m4 := NewStringMatcher(`adj:.:v_zna(?!:rinanim).*`, true, true)
	require.True(t, m4.Matches("adj:m:v_zna"))
	require.True(t, m4.Matches("adj:f:v_zna:pron"))
	require.False(t, m4.Matches("adj:m:v_zna:rinanim"))
}

func TestJavaRegexp_AlternationWithLookahead(t *testing.T) {
	// From UK disambiguation.xml
	pat := `noun:(un)?anim:.:v_zna.*|adj:.:v_zna(?!:rinanim).*|numr:.:v_zna.*`
	m := NewStringMatcher(pat, true, true)
	require.True(t, m.Matches("noun:anim:m:v_zna:prop"))
	require.True(t, m.Matches("adj:f:v_zna"))
	require.False(t, m.Matches("adj:m:v_zna:rinanim"))
	require.True(t, m.Matches("numr:m:v_zna"))
}

func TestJavaRegexp_RE2StillUsedWhenNoLookaround(t *testing.T) {
	// Pattern with no lookaround and not fully enumerable → no javaRE.
	m := NewStringMatcher(`noun:.*x.*`, true, true)
	require.Nil(t, m.javaRE)
	require.True(t, m.Matches("noun:axb"))
	require.False(t, m.Matches("noun:foo"))
}

func TestCompileJavaRegexp_Basic(t *testing.T) {
	jr, err := compileJavaRegexp(`a(?!b)c`, true)
	require.NoError(t, err)
	require.True(t, jr.fullMatch("ac"))
	require.False(t, jr.fullMatch("abc"))
	require.False(t, jr.fullMatch("aXc")) // needs c after a with no b - "aXc" is a + X + c, (?!b) ok at after a
	// a(?!b)c means: 'a', not 'b' at pos, then 'c'. For "aXc": after a, (?!b) succeeds (X!= start of b), then c must match at X — fails.
	require.False(t, jr.fullMatch("aXc"))
}
