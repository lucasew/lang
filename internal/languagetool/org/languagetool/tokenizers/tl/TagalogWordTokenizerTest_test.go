package tl

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.language.tokenizers.TagalogWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// getTokenizingCharacters and inherited WordTokenizer.tokenize with TL delims.

func tlTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestTagalogWordTokenizer_GetTokenizingCharacters(t *testing.T) {
	w := NewTagalogWordTokenizer()
	delims := w.GetTokenizingCharacters()
	base := tokenizers.TokenizingCharacters()

	// Java: super.getTokenizingCharacters() + "-"
	require.True(t, strings.HasPrefix(delims, base), "must include all base WordTokenizer delims as prefix")
	require.Equal(t, base+"-", delims, "exact Java concatenation super + \"-\"")

	// ASCII hyphen-minus is a TL tokenizing character
	require.True(t, strings.ContainsRune(delims, '-'), "ASCII hyphen-minus U+002D is a tokenizing character")

	// suffix exact: "-" (one rune)
	require.True(t, strings.HasSuffix(delims, "-"), "appended \"-\" per Java getTokenizingCharacters")
	last, size := utf8.DecodeLastRuneInString(delims)
	require.NotEqual(t, 0, size, "delims must be non-empty")
	require.NotEqual(t, utf8.RuneError, last)
	require.Equal(t, '-', last, "last rune must be U+002D hyphen-minus")

	// base whitespace delims still present
	require.True(t, strings.ContainsRune(delims, ' '), "ASCII space is base delim")
	require.True(t, strings.ContainsRune(delims, '\u00A0'), "NBSP is base delim")

	// base does NOT include ASCII hyphen-minus (Java comment: not included)
	require.False(t, strings.ContainsRune(base, '-'), "core WordTokenizer base must not include ASCII hyphen-minus")
}

func TestTagalogWordTokenizer_Tokenize(t *testing.T) {
	w := NewTagalogWordTokenizer()

	// ASCII hyphen-minus splits (TL-only among TL delims; base excludes it)
	require.Equal(t, "[well, -, known]", tlTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a, -, b, -, c]", tlTokStr(w.Tokenize("a-b-c")))
	require.Equal(t, "[-]", tlTokStr(w.Tokenize("-")))
	require.Equal(t, "[-, foo]", tlTokStr(w.Tokenize("-foo")))
	require.Equal(t, "[foo, -]", tlTokStr(w.Tokenize("foo-")))
	// multi-hyphen and edges
	require.Equal(t, "[-, -]", tlTokStr(w.Tokenize("--")))
	require.Equal(t, "[a, -, -, b]", tlTokStr(w.Tokenize("a--b")))
	require.Equal(t, "[pre, -, ,, post]", tlTokStr(w.Tokenize("pre-,post")))

	// whitespace / NBSP as base delims (same as core WordTokenizer)
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", tlTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[This, \r, breaks]", tlTokStr(w.Tokenize("This\rbreaks")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[dev.all@languagetool.org]", tlTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", tlTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", tlTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", tlTokStr(w.Tokenize("Mein Adresse address@email.com")))
	// email with hyphen local-part: TL splits on -, join path must reassemble
	require.Equal(t, "[user-name@example.com]", tlTokStr(w.Tokenize("user-name@example.com")))

	// urls joined (same path as core WordTokenizerTest.testUrlTokenize)
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, tlTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[tingnan,  , http://example.com/x]", tlTokStr(w.Tokenize("tingnan http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		tlTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))

	// basic punctuation from base delims still applies
	require.Equal(t, "[Hallo, !,  , Welt, .]", tlTokStr(w.Tokenize("Hallo! Welt.")))
	// ASCII comma still base delim
	require.Equal(t, "[a, ,, b]", tlTokStr(w.Tokenize("a,b")))
	// Tagalog-ish sample with hyphen and space
	require.Equal(t, "[maganda, -, araw,  , po]", tlTokStr(w.Tokenize("maganda-araw po")))
}

// Contrast: core WordTokenizer does NOT split on ASCII hyphen-minus; TL does.
func TestTagalogWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := tokenizers.NewWordTokenizer()
	tl := NewTagalogWordTokenizer()

	// ASCII hyphen: core keeps whole (base excludes -); TL splits
	require.Equal(t, "[well-known]", tlTokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well, -, known]", tlTokStr(tl.Tokenize("well-known")))

	require.Equal(t, "[a-b-c]", tlTokStr(core.Tokenize("a-b-c")))
	require.Equal(t, "[a, -, b, -, c]", tlTokStr(tl.Tokenize("a-b-c")))

	// leading/trailing hyphen
	require.Equal(t, "[-foo]", tlTokStr(core.Tokenize("-foo")))
	require.Equal(t, "[-, foo]", tlTokStr(tl.Tokenize("-foo")))
	require.Equal(t, "[foo-]", tlTokStr(core.Tokenize("foo-")))
	require.Equal(t, "[foo, -]", tlTokStr(tl.Tokenize("foo-")))

	// shared base delims: both split the same on space / punctuation
	require.Equal(t, tlTokStr(core.Tokenize("Hallo! Welt.")), tlTokStr(tl.Tokenize("Hallo! Welt.")))
	require.Equal(t, tlTokStr(core.Tokenize("a,b")), tlTokStr(tl.Tokenize("a,b")))
	require.Equal(t, tlTokStr(core.Tokenize("Das ist\u00A0ein Test")), tlTokStr(tl.Tokenize("Das ist\u00A0ein Test")))
}
