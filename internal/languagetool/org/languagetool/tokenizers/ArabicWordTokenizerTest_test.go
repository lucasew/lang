package tokenizers

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

// Behavior-matrix twin for org.languagetool.tokenizers.ArabicWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// getTokenizingCharacters and inherited WordTokenizer.tokenize with AR delims.

func arTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestArabicWordTokenizer_GetTokenizingCharacters(t *testing.T) {
	w := NewArabicWordTokenizer()
	delims := w.GetTokenizingCharacters()
	base := TokenizingCharacters()

	// Java: super.getTokenizingCharacters() + "،؟؛-"
	require.True(t, strings.HasPrefix(delims, base), "must include all base WordTokenizer delims as prefix")
	require.Equal(t, base+"،؟؛-", delims, "exact Java concatenation super + \"،؟؛-\"")

	// Arabic punctuation + ASCII hyphen-minus are AR tokenizing characters
	require.True(t, strings.ContainsRune(delims, '\u060C'), "Arabic comma U+060C is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '\u061F'), "Arabic question mark U+061F is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '\u061B'), "Arabic semicolon U+061B is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '-'), "ASCII hyphen-minus U+002D is a tokenizing character")

	// suffix exact: "،؟؛-" (four runes)
	require.True(t, strings.HasSuffix(delims, "،؟؛-"), "appended \"،؟؛-\" per Java getTokenizingCharacters")
	// last rune must be ASCII hyphen-minus
	last, size := utf8.DecodeLastRuneInString(delims)
	require.NotEqual(t, 0, size, "delims must be non-empty")
	require.NotEqual(t, utf8.RuneError, last)
	require.Equal(t, '-', last, "last rune must be U+002D hyphen-minus")

	// base whitespace delims still present
	require.True(t, strings.ContainsRune(delims, ' '), "ASCII space is base delim")
	require.True(t, strings.ContainsRune(delims, '\u00A0'), "NBSP is base delim")

	// base does NOT include ASCII hyphen-minus (Java comment: not included)
	require.False(t, strings.ContainsRune(base, '-'), "core WordTokenizer base must not include ASCII hyphen-minus")
	// base does not include the three Arabic punctuation marks
	require.False(t, strings.ContainsRune(base, '\u060C'))
	require.False(t, strings.ContainsRune(base, '\u061F'))
	require.False(t, strings.ContainsRune(base, '\u061B'))
}

func TestArabicWordTokenizer_Tokenize(t *testing.T) {
	w := NewArabicWordTokenizer()

	// Arabic comma ، (U+060C) splits
	require.Equal(t, "[مرحبا, ،,  , عالم]", arTokStr(w.Tokenize("مرحبا، عالم")))
	require.Equal(t, "[هذه, ،, جملة]", arTokStr(w.Tokenize("هذه،جملة")))
	require.Equal(t, "[،]", arTokStr(w.Tokenize("،")))
	require.Equal(t, "[،, هذه]", arTokStr(w.Tokenize("،هذه")))
	require.Equal(t, "[هذه, ،]", arTokStr(w.Tokenize("هذه،")))

	// Arabic question mark ؟ (U+061F) splits
	require.Equal(t, "[مرحبا,  , عالم, ؟]", arTokStr(w.Tokenize("مرحبا عالم؟")))
	require.Equal(t, "[س, ؟, ت]", arTokStr(w.Tokenize("س؟ت")))
	require.Equal(t, "[؟]", arTokStr(w.Tokenize("؟")))

	// Arabic semicolon ؛ (U+061B) splits
	require.Equal(t, "[أ, ؛, ب]", arTokStr(w.Tokenize("أ؛ب")))
	require.Equal(t, "[؛]", arTokStr(w.Tokenize("؛")))
	require.Equal(t, "[أ, ؛,  , ب]", arTokStr(w.Tokenize("أ؛ ب")))

	// ASCII hyphen-minus splits (AR-only among AR delims; base excludes it)
	require.Equal(t, "[well, -, known]", arTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a, -, b, -, c]", arTokStr(w.Tokenize("a-b-c")))
	require.Equal(t, "[-]", arTokStr(w.Tokenize("-")))
	require.Equal(t, "[-, foo]", arTokStr(w.Tokenize("-foo")))
	require.Equal(t, "[foo, -]", arTokStr(w.Tokenize("foo-")))

	// combined AR punctuation
	require.Equal(t, "[مرحبا, ،,  , عالم, ؟]", arTokStr(w.Tokenize("مرحبا، عالم؟")))
	require.Equal(t, "[أ, ،, ب, ؛, ج, ؟]", arTokStr(w.Tokenize("أ،ب؛ج؟")))

	// whitespace / NBSP as base delims (same as core WordTokenizer)
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", arTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[This, \r, breaks]", arTokStr(w.Tokenize("This\rbreaks")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[dev.all@languagetool.org]", arTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", arTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", arTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", arTokStr(w.Tokenize("Mein Adresse address@email.com")))
	// email with hyphen local-part: AR splits on -, join path must reassemble
	require.Equal(t, "[user-name@example.com]", arTokStr(w.Tokenize("user-name@example.com")))

	// urls joined (same path as core WordTokenizerTest.testUrlTokenize)
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, arTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[انظر,  , http://example.com/x]", arTokStr(w.Tokenize("انظر http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		arTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))

	// basic punctuation from base delims still applies
	require.Equal(t, "[Hallo, !,  , Welt, .]", arTokStr(w.Tokenize("Hallo! Welt.")))
	// ASCII comma still base delim
	require.Equal(t, "[a, ,, b]", arTokStr(w.Tokenize("a,b")))
}

// Contrast: core WordTokenizer does NOT split on ، ؟ ؛ or ASCII hyphen-minus.
func TestArabicWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := NewWordTokenizer()
	ar := NewArabicWordTokenizer()

	// Arabic comma: core keeps whole; AR splits
	require.Equal(t, "[هذه،جملة]", arTokStr(core.Tokenize("هذه،جملة")))
	require.Equal(t, "[هذه, ،, جملة]", arTokStr(ar.Tokenize("هذه،جملة")))

	// Arabic question: core keeps whole; AR splits
	require.Equal(t, "[س؟ت]", arTokStr(core.Tokenize("س؟ت")))
	require.Equal(t, "[س, ؟, ت]", arTokStr(ar.Tokenize("س؟ت")))

	// Arabic semicolon: core keeps whole; AR splits
	require.Equal(t, "[أ؛ب]", arTokStr(core.Tokenize("أ؛ب")))
	require.Equal(t, "[أ, ؛, ب]", arTokStr(ar.Tokenize("أ؛ب")))

	// ASCII hyphen: core keeps whole (base excludes -); AR splits
	require.Equal(t, "[well-known]", arTokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well, -, known]", arTokStr(ar.Tokenize("well-known")))
}
