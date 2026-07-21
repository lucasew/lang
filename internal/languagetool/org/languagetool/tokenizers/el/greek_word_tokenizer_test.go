package el

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGreekWordTokenizer(t *testing.T) {
	toks := NewGreekWordTokenizer().Tokenize("Γεια σου")
	require.Equal(t, []string{"Γεια", " ", "σου"}, toks)
}

func TestGreekWordTokenizer_DelimAndSpecial(t *testing.T) {
	// Apostrophe is JFlex Delim (not glued to letters like invent IsLetter path).
	toks := NewGreekWordTokenizer().Tokenize("σ'αγαπώ")
	require.Equal(t, []string{"σ", "'", "αγαπώ"}, toks)

	// NBSP is Delim → own token (unicode.IsSpace invent used to drop it).
	toks = NewGreekWordTokenizer().Tokenize("α\u00A0β")
	require.Equal(t, []string{"α", "\u00A0", "β"}, toks)

	// Special multi-char "ό,τι" (comma would otherwise split).
	toks = NewGreekWordTokenizer().Tokenize("ό,τι άλλο")
	require.Equal(t, []string{"ό,τι", " ", "άλλο"}, toks)

	// Greek Ano Teleia is Delim.
	toks = NewGreekWordTokenizer().Tokenize("α·β")
	require.Equal(t, []string{"α", "·", "β"}, toks)
}

func TestGreekWordTokenizerImpl_RawNoJoin(t *testing.T) {
	// Impl is jflex surface only — raw delims, no email join wrapper needed.
	got := NewGreekWordTokenizerImpl().YylexTokenize("Γεια σου")
	require.Equal(t, []string{"Γεια", " ", "σου"}, got)
}
