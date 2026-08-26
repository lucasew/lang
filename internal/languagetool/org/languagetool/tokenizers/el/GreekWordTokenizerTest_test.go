package el

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.tokenizers.el.GreekWordTokenizer
// (+ tightly coupled GreekWordTokenizerImpl jflex rules).
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes from
// GreekWordTokenizerImpl.jflex Word/Delim + joinEMailsAndUrls.

func elTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

// javaGreekDelimChars is the exact Delim code-point set from
// GreekWordTokenizerImpl.jflex (character-for-character membership).
// Same 73 code points as production greekDelimChars.
const javaGreekDelimChars = "\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}!:\"'" +
	"·" +
	"’‘„“”…«»\\/\t\n"

func TestGreekWordTokenizer_DelimSetExact(t *testing.T) {
	// Production const must match jflex Delim character-for-character.
	require.Equal(t, javaGreekDelimChars, greekDelimChars,
		"greekDelimChars must equal jflex Delim set")

	// Unique code-point count (jflex Delim has 73 distinct code points).
	seen := map[rune]struct{}{}
	for _, r := range greekDelimChars {
		seen[r] = struct{}{}
	}
	require.Equal(t, 73, len(seen), "jflex Delim has 73 distinct code points")
	require.Equal(t, utf8.RuneCountInString(greekDelimChars), len(seen),
		"greekDelimChars must not duplicate code points")

	// Map membership matches the const.
	for _, r := range greekDelimChars {
		require.Truef(t, isGreekDelim(r), "expected delim %U", r)
	}

	// Whitespace / tab / newline / NBSP
	require.True(t, isGreekDelim(' '))
	require.True(t, isGreekDelim('\u00A0'))
	require.True(t, isGreekDelim('\t'))
	require.True(t, isGreekDelim('\n'))

	// Punctuation class from jflex (includes Greek Ano Teleia, guillemets, quotes)
	for _, r := range []rune(",.;()[]{}!:\"'·’‘„“”…«»\\/") {
		require.Truef(t, isGreekDelim(r), "expected delim %U", r)
	}

	// Characters NOT in Greek Delim (contrast tests matter)
	require.False(t, isGreekDelim('?'), "no ? in Greek jflex Delim")
	require.False(t, isGreekDelim('\r'), "no CR in Greek jflex Delim")
	require.False(t, isGreekDelim('-'), "no ASCII hyphen-minus in Greek jflex Delim")
	require.False(t, isGreekDelim('='), "no = in Greek jflex Delim")
	require.False(t, isGreekDelim('\u000B'), "no VT in Greek jflex Delim")
	require.False(t, isGreekDelim('\u2013'), "no en-dash in Greek jflex Delim")

	// Core base WordTokenizer has some of these; Greek jflex does not.
	base := tokenizers.TokenizingCharacters()
	require.True(t, strings.ContainsRune(base, '\r'))
	require.False(t, isGreekDelim('\r'))
	require.True(t, strings.ContainsRune(base, '?'))
	require.False(t, isGreekDelim('?'))
}

func TestGreekWordTokenizer_InheritedGetTokenizingCharactersIsBase(t *testing.T) {
	// Java does not override getTokenizingCharacters — only tokenize uses jflex.
	w := NewGreekWordTokenizer()
	require.Equal(t, tokenizers.TokenizingCharacters(), w.GetTokenizingCharacters(),
		"inherited getTokenizingCharacters must remain base WordTokenizer set")
}

func TestGreekWordTokenizer_Tokenize(t *testing.T) {
	w := NewGreekWordTokenizer()

	// Basic Greek words + space
	require.Equal(t, "[Γεια,  , σου]", elTokStr(w.Tokenize("Γεια σου")))
	require.Equal(t, "[Καλημέρα,  , κόσμε]", elTokStr(w.Tokenize("Καλημέρα κόσμε")))

	// Apostrophe is Delim → splits (not glued)
	require.Equal(t, "[σ, ', αγαπώ]", elTokStr(w.Tokenize("σ'αγαπώ")))

	// NBSP is Delim → own token
	require.Equal(t, "[α, \u00A0, β]", elTokStr(w.Tokenize("α\u00A0β")))
	require.Equal(t, "[a, \t, b]", elTokStr(w.Tokenize("a\tb")))
	require.Equal(t, "[a, \n, b]", elTokStr(w.Tokenize("a\nb")))

	// Special multi-char "ό,τι" (comma would otherwise split)
	require.Equal(t, "[ό,τι]", elTokStr(w.Tokenize("ό,τι")))
	require.Equal(t, "[ό,τι,  , άλλο]", elTokStr(w.Tokenize("ό,τι άλλο")))
	// at start / mid (after delim) / end
	require.Equal(t, "[ό,τι,  , x]", elTokStr(w.Tokenize("ό,τι x")))
	require.Equal(t, "[x,  , ό,τι]", elTokStr(w.Tokenize("x ό,τι")))
	require.Equal(t, "[x,  , ό,τι,  , y]", elTokStr(w.Tokenize("x ό,τι y")))
	// space after comma → not special; splits as ό + , + space + τι
	require.Equal(t, "[ό, ,,  , τι]", elTokStr(w.Tokenize("ό, τι")))
	// two specials back-to-back (each at token start)
	require.Equal(t, "[ό,τι, ό,τι]", elTokStr(w.Tokenize("ό,τιό,τι")))
	// incomplete special → fall back to non-Delim run + Delim pieces
	require.Equal(t, "[ό, ,]", elTokStr(w.Tokenize("ό,")))
	require.Equal(t, "[ό, ,, τ]", elTokStr(w.Tokenize("ό,τ")))
	// special then trailing letter as new token (DFA final after special)
	require.Equal(t, "[ό,τι, ς]", elTokStr(w.Tokenize("ό,τις")))
	// mid-word ό,τι is NOT special: non-Delim run absorbs ό until comma (Delim)
	// Java DFA: "xό" + "," + "τιy"  (special only from lexical state 0)
	require.Equal(t, "[xό, ,, τιy]", elTokStr(w.Tokenize("xό,τιy")))
	require.Equal(t, "[fooό, ,, τιbar]", elTokStr(w.Tokenize("fooό,τιbar")))
	require.Equal(t, "[xxό, ,, τι]", elTokStr(w.Tokenize("xxό,τι")))

	// Greek Ano Teleia U+0387 is Delim
	require.Equal(t, "[α, ·, β]", elTokStr(w.Tokenize("α·β")))
	require.Equal(t, "[·]", elTokStr(w.Tokenize("·")))

	// Each punctuation class that is Delim
	require.Equal(t, "[a, ,, b]", elTokStr(w.Tokenize("a,b")))
	require.Equal(t, "[a, ., b]", elTokStr(w.Tokenize("a.b")))
	require.Equal(t, "[a, ;, b]", elTokStr(w.Tokenize("a;b")))
	require.Equal(t, "[(, x, )]", elTokStr(w.Tokenize("(x)")))
	require.Equal(t, "[[, x, ]]", elTokStr(w.Tokenize("[x]")))
	require.Equal(t, "[{, x, }]", elTokStr(w.Tokenize("{x}")))
	require.Equal(t, "[Hello, !]", elTokStr(w.Tokenize("Hello!")))
	require.Equal(t, "[a, :, b]", elTokStr(w.Tokenize("a:b")))
	require.Equal(t, `["]`, elTokStr(w.Tokenize(`"`)))
	require.Equal(t, "[a, ', b]", elTokStr(w.Tokenize("a'b")))
	require.Equal(t, "[a, ’, b]", elTokStr(w.Tokenize("a’b")))
	require.Equal(t, "[a, ‘, b]", elTokStr(w.Tokenize("a‘b")))
	require.Equal(t, "[a, „, b]", elTokStr(w.Tokenize("a„b")))
	require.Equal(t, "[a, “, b]", elTokStr(w.Tokenize("a“b")))
	require.Equal(t, "[a, ”, b]", elTokStr(w.Tokenize("a”b")))
	require.Equal(t, "[a, …, b]", elTokStr(w.Tokenize("a…b")))
	require.Equal(t, "[«, γεια, »]", elTokStr(w.Tokenize("«γεια»")))
	require.Equal(t, "[a, /, b]", elTokStr(w.Tokenize("a/b")))
	require.Equal(t, `[a, \, b]`, elTokStr(w.Tokenize(`a\b`)))

	// Chars NOT in Delim stay glued
	require.Equal(t, "[a?b]", elTokStr(w.Tokenize("a?b")))
	require.Equal(t, "[well-known]", elTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a-b-c]", elTokStr(w.Tokenize("a-b-c")))
	require.Equal(t, "[This\rbreaks]", elTokStr(w.Tokenize("This\rbreaks")))
	require.Equal(t, "[a=b]", elTokStr(w.Tokenize("a=b")))
	require.Equal(t, "[a–b]", elTokStr(w.Tokenize("a\u2013b")))
	require.Equal(t, "[a\u000Bb]", elTokStr(w.Tokenize("a\u000Bb")))

	// emails/urls joined after jflex split ('.' '/' are Delim → raw split, then join)
	require.Equal(t, "[dev.all@languagetool.org]", elTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", elTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", elTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Η,  , διεύθυνση,  , address@email.com]", elTokStr(w.Tokenize("Η διεύθυνση address@email.com")))
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, elTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[Δες,  , http://example.com/x]", elTokStr(w.Tokenize("Δες http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		elTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens
	require.Empty(t, w.Tokenize(""))
}

func TestGreekWordTokenizerImpl_RawNoJoin(t *testing.T) {
	// Impl is jflex surface only — raw delims, no email/url join.
	impl := NewGreekWordTokenizerImpl()
	require.Equal(t, []string{"Γεια", " ", "σου"}, impl.YylexTokenize("Γεια σου"))
	// raw email pieces before join
	raw := impl.YylexTokenize("dev.all@languagetool.org")
	require.Equal(t, []string{"dev", ".", "all@languagetool", ".", "org"}, raw)
	// Tokenize (wrapper) joins them
	require.Equal(t, []string{"dev.all@languagetool.org"}, NewGreekWordTokenizer().Tokenize("dev.all@languagetool.org"))

	// GetNextToken / GetText loop surface
	impl.Yyreset("α·β")
	require.Equal(t, 0, impl.GetNextToken())
	require.Equal(t, "α", impl.GetText())
	require.Equal(t, 0, impl.GetNextToken())
	require.Equal(t, "·", impl.GetText())
	require.Equal(t, 0, impl.GetNextToken())
	require.Equal(t, "β", impl.GetText())
	require.Equal(t, YYEOF, impl.GetNextToken())

	require.Empty(t, impl.YylexTokenize(""))
}

func TestGreekWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := tokenizers.NewWordTokenizer()
	elw := NewGreekWordTokenizer()

	// Ano Teleia: Greek jflex splits; core base typically does not include U+0387
	require.Equal(t, "[α, ·, β]", elTokStr(elw.Tokenize("α·β")))
	// core keeps whole if · not in base delims
	require.Equal(t, "[α·β]", elTokStr(core.Tokenize("α·β")))

	// '?': core splits; Greek keeps glued
	require.Equal(t, "[a, ?, b]", elTokStr(core.Tokenize("a?b")))
	require.Equal(t, "[a?b]", elTokStr(elw.Tokenize("a?b")))

	// CR: core splits; Greek keeps glued
	require.Equal(t, "[This, \r, breaks]", elTokStr(core.Tokenize("This\rbreaks")))
	require.Equal(t, "[This\rbreaks]", elTokStr(elw.Tokenize("This\rbreaks")))

	// special ό,τι: Greek keeps as one token; core splits on comma
	require.Equal(t, "[ό,τι]", elTokStr(elw.Tokenize("ό,τι")))
	require.Equal(t, "[ό, ,, τι]", elTokStr(core.Tokenize("ό,τι")))

	// shared delims (space, ! .): same split shape for simple ASCII
	require.Equal(t, elTokStr(core.Tokenize("Hallo! Welt.")), elTokStr(elw.Tokenize("Hallo! Welt.")))
}
