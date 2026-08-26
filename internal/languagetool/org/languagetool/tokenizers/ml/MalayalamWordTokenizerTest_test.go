package ml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.tokenizers.ml.MalayalamWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for the
// hardcoded StringTokenizer delims in tokenize(). Java implements Tokenizer
// directly (does NOT extend WordTokenizer) and does NOT call joinEMailsAndUrls.

func mlTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

// Exact Java delimiter string from MalayalamWordTokenizer.tokenize.
const javaMalayalamDelims = "\u0020\u00A0\u115f\u1160\u1680" +
	",.;()[]{}!?:\"'’‘„“”…\\/\t\n"

func TestMalayalamWordTokenizer_DelimStringExact(t *testing.T) {
	// Production const must match Java character-for-character.
	require.Equal(t, javaMalayalamDelims, malayalamDelims, "malayalamDelims must equal Java literal")

	// whitespace / hangul fillers / ogham space / tab / newline from the literal set
	require.True(t, strings.ContainsRune(malayalamDelims, ' '))
	require.True(t, strings.ContainsRune(malayalamDelims, '\u00A0'))
	require.True(t, strings.ContainsRune(malayalamDelims, '\u115f'))
	require.True(t, strings.ContainsRune(malayalamDelims, '\u1160'))
	require.True(t, strings.ContainsRune(malayalamDelims, '\u1680'))
	require.True(t, strings.ContainsRune(malayalamDelims, '\t'))
	require.True(t, strings.ContainsRune(malayalamDelims, '\n'))

	// punctuation subset from the literal set
	for _, r := range []rune(",.;()[]{}!?:\"'’‘„“”…\\/") {
		require.Truef(t, strings.ContainsRune(malayalamDelims, r), "expected delim %U", r)
	}

	// Characters in core base WordTokenizer delims but NOT in ML hardcoded set
	base := tokenizers.TokenizingCharacters()
	require.True(t, strings.ContainsRune(base, '\r'), "core has CR")
	require.False(t, strings.ContainsRune(malayalamDelims, '\r'), "ML delims omit CR")
	require.True(t, strings.ContainsRune(base, '\u000B'), "core has VT")
	require.False(t, strings.ContainsRune(malayalamDelims, '\u000B'), "ML delims omit VT")
	// ASCII hyphen-minus is in neither core base nor ML (Java base comment: not included)
	require.False(t, strings.ContainsRune(base, '-'))
	require.False(t, strings.ContainsRune(malayalamDelims, '-'))
	// en-dash is core-only
	require.True(t, strings.ContainsRune(base, '\u2013'))
	require.False(t, strings.ContainsRune(malayalamDelims, '\u2013'))
	// '=' is core-only
	require.True(t, strings.ContainsRune(base, '='))
	require.False(t, strings.ContainsRune(malayalamDelims, '='))
	// «» are in core (and Khmer) but NOT in ML
	require.True(t, strings.ContainsRune(base, '«'))
	require.True(t, strings.ContainsRune(base, '»'))
	require.False(t, strings.ContainsRune(malayalamDelims, '«'))
	require.False(t, strings.ContainsRune(malayalamDelims, '»'))
	// Khmer signs are Khmer-only (not ML, not base)
	require.False(t, strings.ContainsRune(malayalamDelims, '\u17D4'))
	require.False(t, strings.ContainsRune(malayalamDelims, '\u17D5'))
	// ML includes hangul fillers / ogham space (also in base)
	require.True(t, strings.ContainsRune(base, '\u115f'))
	require.True(t, strings.ContainsRune(base, '\u1160'))
	require.True(t, strings.ContainsRune(base, '\u1680'))
}

func TestMalayalamWordTokenizer_Tokenize(t *testing.T) {
	w := NewMalayalamWordTokenizer()

	// whitespace / NBSP / tab / newline from the literal set
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", mlTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[a, \t, b]", mlTokStr(w.Tokenize("a\tb")))
	require.Equal(t, "[a, \n, b]", mlTokStr(w.Tokenize("a\nb")))
	// hangul fillers / ogham space (in ML delims)
	require.Equal(t, "[a, \u115f, b]", mlTokStr(w.Tokenize("a\u115fb")))
	require.Equal(t, "[a, \u1160, b]", mlTokStr(w.Tokenize("a\u1160b")))
	require.Equal(t, "[a, \u1680, b]", mlTokStr(w.Tokenize("a\u1680b")))
	// CR is NOT in ML delims → must NOT split (contrast core WordTokenizer)
	require.Equal(t, "[This\rbreaks]", mlTokStr(w.Tokenize("This\rbreaks")))

	// punctuation from the literal set
	require.Equal(t, "[a, ,, b]", mlTokStr(w.Tokenize("a,b")))
	require.Equal(t, "[Hallo, !,  , Welt, .]", mlTokStr(w.Tokenize("Hallo! Welt.")))
	require.Equal(t, "[a, ;, b]", mlTokStr(w.Tokenize("a;b")))
	require.Equal(t, "[(, x, )]", mlTokStr(w.Tokenize("(x)")))
	require.Equal(t, "[[, x, ]]", mlTokStr(w.Tokenize("[x]")))
	require.Equal(t, "[{, x, }]", mlTokStr(w.Tokenize("{x}")))
	require.Equal(t, "[a, ?, b, :, c]", mlTokStr(w.Tokenize("a?b:c")))
	require.Equal(t, `["]`, mlTokStr(w.Tokenize(`"`)))
	require.Equal(t, "[']", mlTokStr(w.Tokenize("'")))
	require.Equal(t, "[a, ’, b]", mlTokStr(w.Tokenize("a’b")))
	require.Equal(t, "[a, ‘, b]", mlTokStr(w.Tokenize("a‘b")))
	require.Equal(t, "[a, „, b]", mlTokStr(w.Tokenize("a„b")))
	require.Equal(t, "[a, “, b]", mlTokStr(w.Tokenize("a“b")))
	require.Equal(t, "[a, ”, b]", mlTokStr(w.Tokenize("a”b")))
	require.Equal(t, "[a, …, b]", mlTokStr(w.Tokenize("a…b")))
	require.Equal(t, "[a, /, b]", mlTokStr(w.Tokenize("a/b")))
	require.Equal(t, `[a, \, b]`, mlTokStr(w.Tokenize(`a\b`)))

	// Characters NOT in ML delims but in core base must NOT split under ML
	// ASCII hyphen-minus (neither base nor ML)
	require.Equal(t, "[well-known]", mlTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a-b-c]", mlTokStr(w.Tokenize("a-b-c")))
	// en-dash U+2013: core splits; ML does not
	require.Equal(t, "[a–b]", mlTokStr(w.Tokenize("a\u2013b")))
	// equals: core splits; ML does not
	require.Equal(t, "[a=b]", mlTokStr(w.Tokenize("a=b")))
	// VT: core splits; ML does not
	require.Equal(t, "[a\u000Bb]", mlTokStr(w.Tokenize("a\u000Bb")))
	// «»: core splits; ML does not
	require.Equal(t, "[«x»]", mlTokStr(w.Tokenize("«x»")))
	// Khmer signs: neither core nor ML splits
	require.Equal(t, "[a។b]", mlTokStr(w.Tokenize("a\u17D4b")))
	require.Equal(t, "[a៕b]", mlTokStr(w.Tokenize("a\u17D5b")))

	// emails are NOT re-joined (Java returns raw StringTokenizer output)
	// Contrast: core WordTokenizer / Khmer joinEMailsAndUrls → single token.
	require.Equal(t, "[dev, ., all@languagetool, ., org]", mlTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev, ., all@languagetool, ., org, .]", mlTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev, ., all@languagetool, ., org, :]", mlTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email, ., com]", mlTokStr(w.Tokenize("Mein Adresse address@email.com")))
	// hyphen local-part stays whole under ML delims (no - split); still split on '.'
	require.Equal(t, "[user-name@example, ., com]", mlTokStr(w.Tokenize("user-name@example.com")))

	// urls are NOT re-joined ('.' '/' ':' are delims; no joinEMailsAndUrls)
	require.Equal(t, `[", This,  , http, :, /, /, foo, ., org, ., "]`, mlTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[http, :, /, /, example, ., com, /, x]", mlTokStr(w.Tokenize("http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool, ., org, /, foo, ,,  , and,  , via,  , twitter]",
		mlTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))
	// sole delim
	require.Equal(t, "[ ]", mlTokStr(w.Tokenize(" ")))
	require.Equal(t, "[,]", mlTokStr(w.Tokenize(",")))

	// Malayalam-script sample with space and punctuation
	require.Equal(t, "[നമസ്കാരം, ,,  , ലോകം, .]", mlTokStr(w.Tokenize("നമസ്കാരം, ലോകം.")))
	require.Equal(t, "[hello, ,,  , world]", mlTokStr(w.Tokenize("hello, world")))
}

// Contrast: core WordTokenizer joins emails/urls and splits on a larger delim set;
// ML returns raw tokens on its smaller hardcoded set (no join).
func TestMalayalamWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := tokenizers.NewWordTokenizer()
	ml := NewMalayalamWordTokenizer()

	// CR: core splits; ML keeps whole
	require.Equal(t, "[This, \r, breaks]", mlTokStr(core.Tokenize("This\rbreaks")))
	require.Equal(t, "[This\rbreaks]", mlTokStr(ml.Tokenize("This\rbreaks")))

	// en-dash: core splits; ML keeps whole
	require.Equal(t, "[a, –, b]", mlTokStr(core.Tokenize("a\u2013b")))
	require.Equal(t, "[a–b]", mlTokStr(ml.Tokenize("a\u2013b")))

	// equals: core splits; ML keeps whole
	require.Equal(t, "[a, =, b]", mlTokStr(core.Tokenize("a=b")))
	require.Equal(t, "[a=b]", mlTokStr(ml.Tokenize("a=b")))

	// «»: core splits; ML keeps whole
	require.Equal(t, "[«, x, »]", mlTokStr(core.Tokenize("«x»")))
	require.Equal(t, "[«x»]", mlTokStr(ml.Tokenize("«x»")))

	// shared delims (space, ASCII comma, ! .): both split the same
	require.Equal(t, mlTokStr(core.Tokenize("Hallo! Welt.")), mlTokStr(ml.Tokenize("Hallo! Welt.")))
	require.Equal(t, mlTokStr(core.Tokenize("a,b")), mlTokStr(ml.Tokenize("a,b")))
	require.Equal(t, mlTokStr(core.Tokenize("Das ist\u00A0ein Test")), mlTokStr(ml.Tokenize("Das ist\u00A0ein Test")))

	// ASCII hyphen: both keep whole (neither includes -)
	require.Equal(t, "[well-known]", mlTokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well-known]", mlTokStr(ml.Tokenize("well-known")))

	// emails: core joins; ML leaves raw splits on '.'
	require.Equal(t, "[dev.all@languagetool.org]", mlTokStr(core.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev, ., all@languagetool, ., org]", mlTokStr(ml.Tokenize("dev.all@languagetool.org")))

	// urls: core joins; ML leaves raw splits on ':', '/', '.'
	require.Equal(t, "[http://example.com/x]", mlTokStr(core.Tokenize("http://example.com/x")))
	require.Equal(t, "[http, :, /, /, example, ., com, /, x]", mlTokStr(ml.Tokenize("http://example.com/x")))
}
