package km

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.tokenizers.km.KhmerWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// the hardcoded StringTokenizer delims in tokenize() + joinEMailsAndUrls.
// Java overrides tokenize only — does NOT use getTokenizingCharacters() for
// the tokenize path (inherited getTokenizingCharacters stays base).

func kmTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

// Exact Java delimiter string from KhmerWordTokenizer.tokenize (for membership tests).
const javaKhmerDelims = "\u17D4\u17D5\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}«»!?:\"'’‘„“”…\\/\t\n"

func TestKhmerWordTokenizer_DelimStringExact(t *testing.T) {
	// Production const must match Java character-for-character.
	require.Equal(t, javaKhmerDelims, khmerDelims, "khmerDelims must equal Java literal")

	// Khmer signs present
	require.True(t, strings.ContainsRune(khmerDelims, '\u17D4'), "U+17D4 KHMER SIGN KHAN is a delim")
	require.True(t, strings.ContainsRune(khmerDelims, '\u17D5'), "U+17D5 KHMER SIGN BARIYOOSAN is a delim")
	// first two runes are U+17D4 then U+17D5
	require.True(t, strings.HasPrefix(khmerDelims, "\u17D4\u17D5"))

	// whitespace / tab / newline from the literal set
	require.True(t, strings.ContainsRune(khmerDelims, ' '))
	require.True(t, strings.ContainsRune(khmerDelims, '\u00A0'))
	require.True(t, strings.ContainsRune(khmerDelims, '\t'))
	require.True(t, strings.ContainsRune(khmerDelims, '\n'))

	// punctuation subset from the literal set
	for _, r := range []rune(",.;()[]{}«»!?:\"'’‘„“”…\\/") {
		require.Truef(t, strings.ContainsRune(khmerDelims, r), "expected delim %U", r)
	}

	// Characters in core base WordTokenizer delims but NOT in Khmer hardcoded set
	base := tokenizers.TokenizingCharacters()
	require.True(t, strings.ContainsRune(base, '\r'), "core has CR")
	require.False(t, strings.ContainsRune(khmerDelims, '\r'), "Khmer delims omit CR")
	require.True(t, strings.ContainsRune(base, '\u000B'), "core has VT")
	require.False(t, strings.ContainsRune(khmerDelims, '\u000B'), "Khmer delims omit VT")
	// ASCII hyphen-minus is in neither core base nor Khmer (Java base comment: not included)
	require.False(t, strings.ContainsRune(base, '-'))
	require.False(t, strings.ContainsRune(khmerDelims, '-'))
	// en-dash is core-only
	require.True(t, strings.ContainsRune(base, '\u2013'))
	require.False(t, strings.ContainsRune(khmerDelims, '\u2013'))
	// '=' is core-only
	require.True(t, strings.ContainsRune(base, '='))
	require.False(t, strings.ContainsRune(khmerDelims, '='))
	// U+17D4/U+17D5 are Khmer-only (not base)
	require.False(t, strings.ContainsRune(base, '\u17D4'))
	require.False(t, strings.ContainsRune(base, '\u17D5'))
}

func TestKhmerWordTokenizer_InheritedGetTokenizingCharactersIsBase(t *testing.T) {
	// Java does not override getTokenizingCharacters — only tokenize hardcodes delims.
	w := NewKhmerWordTokenizer()
	require.Equal(t, tokenizers.TokenizingCharacters(), w.GetTokenizingCharacters(),
		"inherited getTokenizingCharacters must remain base WordTokenizer set")
	// Base is NOT the Khmer hardcoded tokenize delims.
	require.NotEqual(t, khmerDelims, w.GetTokenizingCharacters())
}

func TestKhmerWordTokenizer_Tokenize(t *testing.T) {
	w := NewKhmerWordTokenizer()

	// Khmer sign KHAN U+17D4 (។) splits
	require.Equal(t, "[abc, ។, def]", kmTokStr(w.Tokenize("abc\u17D4def")))
	require.Equal(t, "[a, ។, b]", kmTokStr(w.Tokenize("a។b")))
	require.Equal(t, "[។]", kmTokStr(w.Tokenize("។")))
	require.Equal(t, "[។, foo]", kmTokStr(w.Tokenize("។foo")))
	require.Equal(t, "[foo, ។]", kmTokStr(w.Tokenize("foo។")))

	// Khmer sign BARIYOOSAN U+17D5 (៕) splits
	require.Equal(t, "[abc, ៕, def]", kmTokStr(w.Tokenize("abc\u17D5def")))
	require.Equal(t, "[a, ៕, b]", kmTokStr(w.Tokenize("a៕b")))
	require.Equal(t, "[៕]", kmTokStr(w.Tokenize("៕")))
	require.Equal(t, "[៕, foo]", kmTokStr(w.Tokenize("៕foo")))
	require.Equal(t, "[foo, ៕]", kmTokStr(w.Tokenize("foo៕")))

	// both Khmer signs
	require.Equal(t, "[x, ។, y, ៕, z]", kmTokStr(w.Tokenize("x។y៕z")))
	require.Equal(t, "[។, ៕]", kmTokStr(w.Tokenize("។៕")))

	// whitespace / NBSP / tab / newline from the literal set
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", kmTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[a, \t, b]", kmTokStr(w.Tokenize("a\tb")))
	require.Equal(t, "[a, \n, b]", kmTokStr(w.Tokenize("a\nb")))
	// CR is NOT in Khmer delims → must NOT split (contrast core WordTokenizer)
	require.Equal(t, "[This\rbreaks]", kmTokStr(w.Tokenize("This\rbreaks")))

	// punctuation from the literal set
	require.Equal(t, "[a, ,, b]", kmTokStr(w.Tokenize("a,b")))
	require.Equal(t, "[Hallo, !,  , Welt, .]", kmTokStr(w.Tokenize("Hallo! Welt.")))
	require.Equal(t, "[a, ;, b]", kmTokStr(w.Tokenize("a;b")))
	require.Equal(t, "[(, x, )]", kmTokStr(w.Tokenize("(x)")))
	require.Equal(t, "[[, x, ]]", kmTokStr(w.Tokenize("[x]")))
	require.Equal(t, "[{, x, }]", kmTokStr(w.Tokenize("{x}")))
	require.Equal(t, "[«, x, »]", kmTokStr(w.Tokenize("«x»")))
	require.Equal(t, "[a, ?, b, :, c]", kmTokStr(w.Tokenize("a?b:c")))
	require.Equal(t, `["]`, kmTokStr(w.Tokenize(`"`)))
	require.Equal(t, "[a, …, b]", kmTokStr(w.Tokenize("a…b")))
	require.Equal(t, "[a, /, b]", kmTokStr(w.Tokenize("a/b")))
	require.Equal(t, `[a, \, b]`, kmTokStr(w.Tokenize(`a\b`)))

	// Characters NOT in Khmer delims but in core base must NOT split under Khmer
	// ASCII hyphen-minus (neither base nor Khmer)
	require.Equal(t, "[well-known]", kmTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a-b-c]", kmTokStr(w.Tokenize("a-b-c")))
	// en-dash U+2013: core splits; Khmer does not
	require.Equal(t, "[a–b]", kmTokStr(w.Tokenize("a\u2013b")))
	// equals: core splits; Khmer does not
	require.Equal(t, "[a=b]", kmTokStr(w.Tokenize("a=b")))
	// VT: core splits; Khmer does not
	require.Equal(t, "[a\u000Bb]", kmTokStr(w.Tokenize("a\u000Bb")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[dev.all@languagetool.org]", kmTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", kmTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", kmTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", kmTokStr(w.Tokenize("Mein Adresse address@email.com")))
	// hyphen local-part stays whole under Khmer delims (no - split); still valid email
	require.Equal(t, "[user-name@example.com]", kmTokStr(w.Tokenize("user-name@example.com")))

	// urls joined (same path as core WordTokenizerTest.testUrlTokenize)
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, kmTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[មើល,  , http://example.com/x]", kmTokStr(w.Tokenize("មើល http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		kmTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))

	// Khmer-ish sample with space and KHAN
	require.Equal(t, "[សួស្តី,  , ពិភពលោក, ។]", kmTokStr(w.Tokenize("សួស្តី ពិភពលោក។")))
}

// Contrast: core WordTokenizer does NOT split on U+17D4/U+17D5; Khmer does.
// Core splits on CR / en-dash / '='; Khmer does not (hardcoded subset).
func TestKhmerWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := tokenizers.NewWordTokenizer()
	km := NewKhmerWordTokenizer()

	// U+17D4 KHAN: core keeps whole; Khmer splits
	require.Equal(t, "[abc។def]", kmTokStr(core.Tokenize("abc\u17D4def")))
	require.Equal(t, "[abc, ។, def]", kmTokStr(km.Tokenize("abc\u17D4def")))

	// U+17D5 BARIYOOSAN: core keeps whole; Khmer splits
	require.Equal(t, "[abc៕def]", kmTokStr(core.Tokenize("abc\u17D5def")))
	require.Equal(t, "[abc, ៕, def]", kmTokStr(km.Tokenize("abc\u17D5def")))

	// CR: core splits; Khmer keeps whole
	require.Equal(t, "[This, \r, breaks]", kmTokStr(core.Tokenize("This\rbreaks")))
	require.Equal(t, "[This\rbreaks]", kmTokStr(km.Tokenize("This\rbreaks")))

	// en-dash: core splits; Khmer keeps whole
	require.Equal(t, "[a, –, b]", kmTokStr(core.Tokenize("a\u2013b")))
	require.Equal(t, "[a–b]", kmTokStr(km.Tokenize("a\u2013b")))

	// equals: core splits; Khmer keeps whole
	require.Equal(t, "[a, =, b]", kmTokStr(core.Tokenize("a=b")))
	require.Equal(t, "[a=b]", kmTokStr(km.Tokenize("a=b")))

	// shared delims (space, ASCII comma, ! .): both split the same
	require.Equal(t, kmTokStr(core.Tokenize("Hallo! Welt.")), kmTokStr(km.Tokenize("Hallo! Welt.")))
	require.Equal(t, kmTokStr(core.Tokenize("a,b")), kmTokStr(km.Tokenize("a,b")))
	require.Equal(t, kmTokStr(core.Tokenize("Das ist\u00A0ein Test")), kmTokStr(km.Tokenize("Das ist\u00A0ein Test")))

	// ASCII hyphen: both keep whole (neither includes -)
	require.Equal(t, "[well-known]", kmTokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well-known]", kmTokStr(km.Tokenize("well-known")))
}
