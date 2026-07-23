package ru

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Behavior-matrix twin for org.languagetool.tokenizers.ru.RussianWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for every branch.

func tokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestRussianWordTokenizer_GetTokenizingCharacters(t *testing.T) {
	w := NewRussianWordTokenizer()
	delims := w.GetTokenizingCharacters()
	// Java: super.getTokenizingCharacters() + "'."
	require.True(t, strings.ContainsRune(delims, '\''), "apostrophe is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '.'), "period is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '/'), "slash remains base delimiter (б/у needs protect)")
	require.True(t, strings.HasSuffix(delims, "'."), "appended \"'.\" per Java getTokenizingCharacters")
}

func TestRussianWordTokenizer_Tokenize(t *testing.T) {
	w := NewRussianWordTokenizer()

	// б/у stays whole (not split on /)
	require.Equal(t, "[купить,  , б/у,  , телефон]", tokStr(w.Tokenize("купить б/у телефон")))
	require.Equal(t, "[б/у]", tokStr(w.Tokenize("б/у")))

	// б/н stays whole
	require.Equal(t, "[оплата,  , б/н]", tokStr(w.Tokenize("оплата б/н")))
	require.Equal(t, "[б/н]", tokStr(w.Tokenize("б/н")))

	// both abbreviations in one string
	require.Equal(t, "[б/у,  , б/н]", tokStr(w.Tokenize("б/у б/н")))

	// other slash compounds still split (only б/у and б/н are protected)
	require.Equal(t, "[км, /, ч]", tokStr(w.Tokenize("км/ч")))

	// period is its own token (delimiter from super + "'.")
	require.Equal(t, "[слово, ., слово]", tokStr(w.Tokenize("слово.слово")))

	// apostrophe is its own token
	require.Equal(t, "[кто, ', то]", tokStr(w.Tokenize("кто'то")))

	// ASCII hyphen is not a base delimiter → compound stays whole
	require.Equal(t, "[кто-то]", tokStr(w.Tokenize("кто-то")))

	// trailing " ." → space + "." tokens (SP_DOT path; not glued to previous word)
	require.Equal(t, "[конец,  , .]", tokStr(w.Tokenize("конец .")))
	require.Equal(t, "[foo,  , .]", tokStr(w.Tokenize("foo .")))

	// mid " . " restored then normal split
	require.Equal(t, "[a,  , .,  , b]", tokStr(w.Tokenize("a . b")))

	// " .. " path (protected so trailing " ." replace does not corrupt, then restored)
	require.Equal(t, "[a,  , ., .,  , b]", tokStr(w.Tokenize("a .. b")))

	// combined space-dot patterns
	require.Equal(t, "[x,  , .,  , y,  , ., .,  , z]", tokStr(w.Tokenize("x . y .. z")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[Мой,  , адрес,  , address@email.com]", tokStr(w.Tokenize("Мой адрес address@email.com")))
	require.Equal(t, "[dev.all@languagetool.org]", tokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", tokStr(w.Tokenize("dev.all@languagetool.org.")))

	// urls joined
	require.Equal(t, "[см, .,  , http://example.com/x]", tokStr(w.Tokenize("см. http://example.com/x")))
	// same URL-join path as core WordTokenizerTest.testUrlTokenize
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, tokStr(w.Tokenize(`"This http://foo.org."`)))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))
}
