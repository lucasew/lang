package gl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.tokenizers.gl.GalicianWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// pre-protect patterns, SPLIT_CHARS StringTokenizer path, restore, joinEMailsAndUrls.
// Reference outcomes checked against a local Java transcription of tokenize()
// (without invent rewrites of DATE alt2/alt3 or DECIMAL_SPACE lookaround).

func glTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

// Exact Java GalicianWordTokenizer.SPLIT_CHARS (for membership / equality tests).
const javaGLSplitChars = "\u0020\u002d\u00A0" +
	"\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008" +
	"\u2009\u2013\u2014\u2015\u200A\u200B\u200c\u200d\u200e" +
	"\u200f\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	"\u002A\u002B×∗·÷:=≠≂≃≄≅≆≇≈≉≤≥≪≫∧∨∩∪∈∉∊∋∌∍" +
	",.;<>()[]{}¿¡!?:\"«»`'’‘„“”…\\/\t\r\n"

func TestGalicianWordTokenizer_SplitCharsExact(t *testing.T) {
	require.Equal(t, javaGLSplitChars, splitChars, "splitChars must equal Java SPLIT_CHARS literal")

	// ASCII hyphen-minus is a Galician split char (unlike base WordTokenizer)
	require.True(t, strings.ContainsRune(splitChars, '-'))
	require.True(t, strings.ContainsRune(splitChars, '\u002D'))
	// en/em dashes
	require.True(t, strings.ContainsRune(splitChars, '\u2013'))
	require.True(t, strings.ContainsRune(splitChars, '\u2014'))
	// ¿ ¡
	require.True(t, strings.ContainsRune(splitChars, '¿'))
	require.True(t, strings.ContainsRune(splitChars, '¡'))
	// whitespace family from literal
	require.True(t, strings.ContainsRune(splitChars, ' '))
	require.True(t, strings.ContainsRune(splitChars, '\u00A0'))
	require.True(t, strings.ContainsRune(splitChars, '\t'))
	require.True(t, strings.ContainsRune(splitChars, '\r'))
	require.True(t, strings.ContainsRune(splitChars, '\n'))
	// math ops / equals / colon (colon also protected when digit:digit)
	require.True(t, strings.ContainsRune(splitChars, '='))
	require.True(t, strings.ContainsRune(splitChars, ':'))
	require.True(t, strings.ContainsRune(splitChars, '×'))
	require.True(t, strings.ContainsRune(splitChars, '÷'))
}

func TestGalicianWordTokenizer_InheritedGetTokenizingCharactersIsBase(t *testing.T) {
	// Java does not override getTokenizingCharacters — only tokenize uses SPLIT_CHARS.
	// GalicianWordTokenizer has no GetTokenizingCharacters method; base set still differs.
	base := tokenizers.TokenizingCharacters()
	require.NotEqual(t, splitChars, base, "Galician SPLIT_CHARS is not the base WordTokenizer set")
	require.False(t, strings.ContainsRune(base, '-'), "base omits ASCII hyphen-minus")
	require.True(t, strings.ContainsRune(splitChars, '-'))
}

func TestGalicianWordTokenizer_DecimalComma(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// decimal comma between digits stays one token
	require.Equal(t, "[3,14]", glTokStr(w.Tokenize("3,14")))
	require.Equal(t, "[1.234,56]", glTokStr(w.Tokenize("1.234,56")))
	// non-digit comma still splits
	require.Equal(t, "[a, ,, b]", glTokStr(w.Tokenize("a,b")))
	require.Equal(t, "[3, ,, x]", glTokStr(w.Tokenize("3,x")))
}

func TestGalicianWordTokenizer_DottedNumbersAndDates(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// dotted number protected when '.' not last char
	require.Equal(t, "[1.2]", glTokStr(w.Tokenize("1.2")))
	require.Equal(t, "[1.234,56]", glTokStr(w.Tokenize("1.234,56")))
	// dd.MM.yyyy (DATE alt1) — $1 $2 $3 filled → protected
	require.Equal(t, "[12.03.2020]", glTokStr(w.Tokenize("12.03.2020")))
	require.Equal(t, "[xx,  , 12.03.2020,  , yy]", glTokStr(w.Tokenize("xx 12.03.2020 yy")))
	// yyyy.MM.dd (DATE alt2) — groups 1–3 empty; Java REPL leaves ".." as one token
	require.Equal(t, "[..]", glTokStr(w.Tokenize("2020.03.12")))
	require.Equal(t, "[xx,  , ..,  , yy]", glTokStr(w.Tokenize("xx 2020.03.12 yy")))
	// yyyy-MM-dd alone: no '.' in text → DATE gate closed → hyphen splits
	require.Equal(t, "[2020, -, 03, -, 12]", glTokStr(w.Tokenize("2020-03-12")))
	require.Equal(t, "[xx,  , 2020, -, 03, -, 12,  , yy]", glTokStr(w.Tokenize("xx 2020-03-12 yy")))
	// yyyy-MM-dd with a non-trailing '.' elsewhere: DATE runs; alt3 → ".."
	require.Equal(t, "[xx,  , .., .,  , yy]", glTokStr(w.Tokenize("xx 2020-03-12. yy")))
}

func TestGalicianWordTokenizer_DottedOrdinals(t *testing.T) {
	w := NewGalicianWordTokenizer()
	require.Equal(t, "[1.º]", glTokStr(w.Tokenize("1.º")))
	require.Equal(t, "[12.a]", glTokStr(w.Tokenize("12.a")))
	require.Equal(t, "[1.as]", glTokStr(w.Tokenize("1.as")))
	require.Equal(t, "[6.as]", glTokStr(w.Tokenize("6.as")))
	require.Equal(t, "[4.ª]", glTokStr(w.Tokenize("4.ª")))
	require.Equal(t, "[5.ºˢ]", glTokStr(w.Tokenize("5.ºˢ")))
	// CASE_INSENSITIVE|UNICODE_CASE
	require.Equal(t, "[2.O]", glTokStr(w.Tokenize("2.O")))
	require.Equal(t, "[3.A]", glTokStr(w.Tokenize("3.A")))
	// trailing sentence period after ordinal
	require.Equal(t, "[1.º, .]", glTokStr(w.Tokenize("1.º.")))
}

func TestGalicianWordTokenizer_DotInsideSentenceGate(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// first '.' is last character → no DATE/DOTTED_NUMBERS/ORDINALS protection
	require.Equal(t, "[1, .]", glTokStr(w.Tokenize("1.")))
	require.Equal(t, "[Ends,  , 1, .]", glTokStr(w.Tokenize("Ends 1.")))
	// '.' not last → protect digit.digit
	require.Equal(t, "[1.2]", glTokStr(w.Tokenize("1.2")))
}

func TestGalicianWordTokenizer_SpacedThousands(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// Java DECIMAL_SPACE: (?<=^|[\s(])\d{1,3}( [\d]{3})+(?=[\s(]|$)
	require.Equal(t, "[2 000 000]", glTokStr(w.Tokenize("2 000 000")))
	require.Equal(t, "[x,  , 2 000 000,  , y]", glTokStr(w.Tokenize("x 2 000 000 y")))
	require.Equal(t, "[ , 2 000 000]", glTokStr(w.Tokenize(" 2 000 000")))
	require.Equal(t, "[2 000 000,  ]", glTokStr(w.Tokenize("2 000 000 ")))
	require.Equal(t, "[12 000]", glTokStr(w.Tokenize("12 000")))
	// right boundary not [\s(]|$ → not fully protected; Java backtracks to shorter group
	require.Equal(t, "[2 000,  , 000x]", glTokStr(w.Tokenize("2 000 000x")))
	// right is '.' — not in Galician lookaround (unlike PT \D)
	require.Equal(t, "[2,  , 000, .]", glTokStr(w.Tokenize("2 000.")))
	// right is ',' — not protected
	require.Equal(t, "[2,  , 000, ,]", glTokStr(w.Tokenize("2 000,")))
	// left not boundary
	require.Equal(t, "[a1,  , 000,  , b]", glTokStr(w.Tokenize("a1 000 b")))
	// only one space group of 3 digits required; "1 23" no match
	require.Equal(t, "[1,  , 23]", glTokStr(w.Tokenize("1 23")))
	// \d{1,3} cannot start a 4-digit head
	require.Equal(t, "[1000,  , 000]", glTokStr(w.Tokenize("1000 000")))
	// parens: right ')' fails; with trailing space inside parens OK
	require.Equal(t, "[(, 2,  , 000, )]", glTokStr(w.Tokenize("(2 000)")))
	require.Equal(t, "[(, 2 000,  , )]", glTokStr(w.Tokenize("(2 000 )")))
}

func TestGalicianWordTokenizer_ColonNumbers(t *testing.T) {
	w := NewGalicianWordTokenizer()
	require.Equal(t, "[12:25]", glTokStr(w.Tokenize("12:25")))
	require.Equal(t, "[12:25:30]", glTokStr(w.Tokenize("12:25:30")))
	// non-digit colon still splits
	require.Equal(t, "[a, :, b]", glTokStr(w.Tokenize("a:b")))
}

func TestGalicianWordTokenizer_HyphenAndDashes(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// ASCII hyphen is in SPLIT_CHARS
	require.Equal(t, "[pre, -, escolar]", glTokStr(w.Tokenize("pre-escolar")))
	require.Equal(t, "[a, -, b, -, c]", glTokStr(w.Tokenize("a-b-c")))
	// en / em dashes
	require.Equal(t, "[a, –, b]", glTokStr(w.Tokenize("a\u2013b")))
	require.Equal(t, "[a, —, b]", glTokStr(w.Tokenize("a\u2014b")))
}

func TestGalicianWordTokenizer_BasicGalicianPunct(t *testing.T) {
	w := NewGalicianWordTokenizer()
	require.Equal(t, "[Ola, ,,  , mundo, !]", glTokStr(w.Tokenize("Ola, mundo!")))
	require.Equal(t, "[¿, Qué, ?]", glTokStr(w.Tokenize("¿Qué?")))
	require.Equal(t, "[¡, Ola, !]", glTokStr(w.Tokenize("¡Ola!")))
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", glTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[This, \r, breaks]", glTokStr(w.Tokenize("This\rbreaks")))
	require.Equal(t, "[a, \t, b]", glTokStr(w.Tokenize("a\tb")))
	require.Equal(t, "[a, \n, b]", glTokStr(w.Tokenize("a\nb")))
	require.Empty(t, w.Tokenize(""))
	require.Equal(t, "[ ]", glTokStr(w.Tokenize(" ")))
}

func TestGalicianWordTokenizer_EmailsAndUrls(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// joinEMailsAndUrls after split (hyphen in SPLIT_CHARS still re-joined for emails)
	require.Equal(t, "[dev.all@languagetool.org]", glTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", glTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", glTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", glTokStr(w.Tokenize("Mein Adresse address@email.com")))
	require.Equal(t, "[user-name@example.com]", glTokStr(w.Tokenize("user-name@example.com")))
	// urls
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, glTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		glTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))
}

func TestGalicianWordTokenizer_CombinedProtections(t *testing.T) {
	w := NewGalicianWordTokenizer()
	// comma + dotted number both protected
	require.Equal(t, "[3,14.5]", glTokStr(w.Tokenize("3,14.5")))
	// colon + space context
	require.Equal(t, "[Son,  , as,  , 12:25,  , en,  , Vigo, .]", glTokStr(w.Tokenize("Son as 12:25 en Vigo.")))
}
