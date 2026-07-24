package de

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Behavior-matrix twin for org.languagetool.tokenizers.de.GermanWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// getTokenizingCharacters and inherited WordTokenizer.tokenize with DE delims.

func tokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestGermanWordTokenizer_GetTokenizingCharacters(t *testing.T) {
	w := NewGermanWordTokenizer()
	delims := w.GetTokenizingCharacters()
	base := tokenizers.TokenizingCharacters()

	// Java: super.getTokenizingCharacters() + "_‚"
	require.True(t, strings.HasPrefix(delims, base), "must include all base WordTokenizer delims as prefix")
	require.Equal(t, base+"_‚", delims, "exact Java concatenation super + \"_‚\"")

	// underscore is a DE tokenizing character
	require.True(t, strings.ContainsRune(delims, '_'), "underscore is a tokenizing character")

	// ‚ is U+201A single low-9 quotation mark (not ASCII comma U+002C)
	require.True(t, strings.HasSuffix(delims, "_‚"), "appended \"_‚\" per Java getTokenizingCharacters")
	last, size := utf8.DecodeLastRuneInString(delims)
	require.NotEqual(t, 0, size, "delims must be non-empty")
	require.NotEqual(t, utf8.RuneError, last)
	require.Equal(t, '\u201A', last, "last rune must be U+201A low-9 quote, not comma")
	require.NotEqual(t, ',', last)

	// base whitespace delims still present
	require.True(t, strings.ContainsRune(delims, ' '), "ASCII space is base delim")
	require.True(t, strings.ContainsRune(delims, '\u00A0'), "NBSP is base delim")
}

func TestGermanWordTokenizer_Tokenize(t *testing.T) {
	w := NewGermanWordTokenizer()

	// underscore splits (DE-only delim): foo_bar → [foo, _, bar]
	require.Equal(t, "[foo, _, bar]", tokStr(w.Tokenize("foo_bar")))
	require.Equal(t, "[a, _, b, _, c]", tokStr(w.Tokenize("a_b_c")))
	require.Equal(t, "[_]", tokStr(w.Tokenize("_")))
	require.Equal(t, "[_, foo]", tokStr(w.Tokenize("_foo")))
	require.Equal(t, "[foo, _]", tokStr(w.Tokenize("foo_")))

	// low-9 quotation mark ‚ (U+201A) splits
	require.Equal(t, "[sagte, ‚, hallo]", tokStr(w.Tokenize("sagte‚hallo")))
	require.Equal(t, "[‚]", tokStr(w.Tokenize("‚")))
	require.Equal(t, "[er,  , sagte, ‚, hallo, ‚]", tokStr(w.Tokenize("er sagte‚hallo‚")))

	// ASCII comma is base delim (contrast with low-9): still splits
	require.Equal(t, "[a, ,, b]", tokStr(w.Tokenize("a,b")))

	// whitespace / NBSP as base delims (same as core WordTokenizer)
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", tokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[This, \r, breaks]", tokStr(w.Tokenize("This\rbreaks")))

	// ASCII hyphen is NOT a delim (base comment: not included) — not DE-special
	require.Equal(t, "[well-known]", tokStr(w.Tokenize("well-known")))
	require.Equal(t, "[Nord-Süd-Bahn]", tokStr(w.Tokenize("Nord-Süd-Bahn")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[dev.all@languagetool.org]", tokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", tokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", tokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", tokStr(w.Tokenize("Mein Adresse address@email.com")))
	// email with underscore local-part: DE splits on _, so join path still must reassemble
	// Java E_MAIL can match across concatenated tokens after split; joinEMails rejoins.
	require.Equal(t, "[user_name@example.com]", tokStr(w.Tokenize("user_name@example.com")))

	// urls joined (same path as core WordTokenizerTest.testUrlTokenize)
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, tokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[siehe,  , http://example.com/x]", tokStr(w.Tokenize("siehe http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		tokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))

	// basic punctuation from base delims still applies
	require.Equal(t, "[Hallo, !,  , Welt, .]", tokStr(w.Tokenize("Hallo! Welt.")))
}

// Contrast: core WordTokenizer does NOT split on underscore or low-9 quote.
func TestGermanWordTokenizer_ContrastWithCoreWordTokenizer(t *testing.T) {
	core := tokenizers.NewWordTokenizer()
	de := NewGermanWordTokenizer()

	// underscore: core keeps whole; DE splits
	require.Equal(t, "[foo_bar]", tokStr(core.Tokenize("foo_bar")))
	require.Equal(t, "[foo, _, bar]", tokStr(de.Tokenize("foo_bar")))

	// low-9: core keeps whole; DE splits
	require.Equal(t, "[sagte‚hallo]", tokStr(core.Tokenize("sagte‚hallo")))
	require.Equal(t, "[sagte, ‚, hallo]", tokStr(de.Tokenize("sagte‚hallo")))

	// ASCII hyphen: both keep whole (not a base delim; DE does not special-case)
	require.Equal(t, "[well-known]", tokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well-known]", tokStr(de.Tokenize("well-known")))
}
