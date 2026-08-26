package tokenizers

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

// Behavior-matrix twin for org.languagetool.tokenizers.PersianWordTokenizer.
// No dedicated Java *WordTokenizerTest; asserts Java-visible outcomes for
// getTokenizingCharacters and inherited WordTokenizer.tokenize with FA delims.

func faTokStr(tokens []string) string {
	// Java List.toString()
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestPersianWordTokenizer_GetTokenizingCharacters(t *testing.T) {
	w := NewPersianWordTokenizer()
	delims := w.GetTokenizingCharacters()
	base := TokenizingCharacters()

	// Java: super.getTokenizingCharacters() + "،؟؛"
	require.True(t, strings.HasPrefix(delims, base), "must include all base WordTokenizer delims as prefix")
	require.Equal(t, base+"،؟؛", delims, "exact Java concatenation super + \"،؟؛\"")

	// Persian/Arabic punctuation are FA tokenizing characters
	require.True(t, strings.ContainsRune(delims, '\u060C'), "Arabic comma U+060C is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '\u061F'), "Arabic question mark U+061F is a tokenizing character")
	require.True(t, strings.ContainsRune(delims, '\u061B'), "Arabic semicolon U+061B is a tokenizing character")

	// suffix exact: "،؟؛" (three runes) — NO ASCII hyphen (contrast ArabicWordTokenizer)
	require.True(t, strings.HasSuffix(delims, "،؟؛"), "appended \"،؟؛\" per Java getTokenizingCharacters")
	last, size := utf8.DecodeLastRuneInString(delims)
	require.NotEqual(t, 0, size, "delims must be non-empty")
	require.NotEqual(t, utf8.RuneError, last)
	require.Equal(t, '\u061B', last, "last rune must be U+061B Arabic semicolon")
	require.False(t, strings.ContainsRune(delims, '-'), "FA must NOT include ASCII hyphen-minus (contrast AR)")

	// base whitespace delims still present
	require.True(t, strings.ContainsRune(delims, ' '), "ASCII space is base delim")
	require.True(t, strings.ContainsRune(delims, '\u00A0'), "NBSP is base delim")

	// base does not include the three Arabic punctuation marks
	require.False(t, strings.ContainsRune(base, '\u060C'))
	require.False(t, strings.ContainsRune(base, '\u061F'))
	require.False(t, strings.ContainsRune(base, '\u061B'))
	// base does NOT include ASCII hyphen-minus (Java comment: not included)
	require.False(t, strings.ContainsRune(base, '-'), "core WordTokenizer base must not include ASCII hyphen-minus")
}

func TestPersianWordTokenizer_Tokenize(t *testing.T) {
	w := NewPersianWordTokenizer()

	// Arabic comma ، (U+060C) splits
	require.Equal(t, "[سلام, ،,  , دنیا]", faTokStr(w.Tokenize("سلام، دنیا")))
	require.Equal(t, "[این, ،, جمله]", faTokStr(w.Tokenize("این،جمله")))
	require.Equal(t, "[،]", faTokStr(w.Tokenize("،")))
	require.Equal(t, "[،, این]", faTokStr(w.Tokenize("،این")))
	require.Equal(t, "[این, ،]", faTokStr(w.Tokenize("این،")))

	// Arabic question mark ؟ (U+061F) splits
	require.Equal(t, "[سلام,  , دنیا, ؟]", faTokStr(w.Tokenize("سلام دنیا؟")))
	require.Equal(t, "[س, ؟, ت]", faTokStr(w.Tokenize("س؟ت")))
	require.Equal(t, "[؟]", faTokStr(w.Tokenize("؟")))

	// Arabic semicolon ؛ (U+061B) splits
	require.Equal(t, "[أ, ؛, ب]", faTokStr(w.Tokenize("أ؛ب")))
	require.Equal(t, "[؛]", faTokStr(w.Tokenize("؛")))
	require.Equal(t, "[أ, ؛,  , ب]", faTokStr(w.Tokenize("أ؛ ب")))

	// ASCII hyphen-minus does NOT split (FA omits -; contrast AR)
	require.Equal(t, "[well-known]", faTokStr(w.Tokenize("well-known")))
	require.Equal(t, "[a-b-c]", faTokStr(w.Tokenize("a-b-c")))
	require.Equal(t, "[-]", faTokStr(w.Tokenize("-")))
	require.Equal(t, "[-foo]", faTokStr(w.Tokenize("-foo")))
	require.Equal(t, "[foo-]", faTokStr(w.Tokenize("foo-")))

	// combined FA punctuation
	require.Equal(t, "[سلام, ،,  , دنیا, ؟]", faTokStr(w.Tokenize("سلام، دنیا؟")))
	require.Equal(t, "[أ, ،, ب, ؛, ج, ؟]", faTokStr(w.Tokenize("أ،ب؛ج؟")))

	// whitespace / NBSP as base delims (same as core WordTokenizer)
	require.Equal(t, "[Das,  , ist, \u00A0, ein,  , Test]", faTokStr(w.Tokenize("Das ist\u00A0ein Test")))
	require.Equal(t, "[This, \r, breaks]", faTokStr(w.Tokenize("This\rbreaks")))

	// emails joined like core WordTokenizer.joinEMailsAndUrls
	require.Equal(t, "[dev.all@languagetool.org]", faTokStr(w.Tokenize("dev.all@languagetool.org")))
	require.Equal(t, "[dev.all@languagetool.org, .]", faTokStr(w.Tokenize("dev.all@languagetool.org.")))
	require.Equal(t, "[dev.all@languagetool.org, :]", faTokStr(w.Tokenize("dev.all@languagetool.org:")))
	require.Equal(t, "[Mein,  , Adresse,  , address@email.com]", faTokStr(w.Tokenize("Mein Adresse address@email.com")))
	// hyphen local-part stays whole under FA delims (no - split); still valid email
	require.Equal(t, "[user-name@example.com]", faTokStr(w.Tokenize("user-name@example.com")))

	// urls joined (same path as core WordTokenizerTest.testUrlTokenize)
	require.Equal(t, `[", This,  , http://foo.org, ., "]`, faTokStr(w.Tokenize(`"This http://foo.org."`)))
	require.Equal(t, "[ببین,  , http://example.com/x]", faTokStr(w.Tokenize("ببین http://example.com/x")))
	require.Equal(t, "[Get,  , more,  , at,  , languagetool.org/foo, ,,  , and,  , via,  , twitter]",
		faTokStr(w.Tokenize("Get more at languagetool.org/foo, and via twitter")))

	// empty input → no tokens (StringTokenizer empty)
	require.Empty(t, w.Tokenize(""))

	// basic punctuation from base delims still applies
	require.Equal(t, "[Hallo, !,  , Welt, .]", faTokStr(w.Tokenize("Hallo! Welt.")))
	// ASCII comma still base delim
	require.Equal(t, "[a, ,, b]", faTokStr(w.Tokenize("a,b")))
}

// Contrast: core WordTokenizer does NOT split on ، ؟ ؛; FA does. Hyphen stays whole for both.
// Contrast AR: ArabicWordTokenizer also splits on ASCII hyphen; FA does not.
func TestPersianWordTokenizer_ContrastWithCoreAndArabic(t *testing.T) {
	core := NewWordTokenizer()
	fa := NewPersianWordTokenizer()
	ar := NewArabicWordTokenizer()

	// Arabic comma: core keeps whole; FA splits
	require.Equal(t, "[این،جمله]", faTokStr(core.Tokenize("این،جمله")))
	require.Equal(t, "[این, ،, جمله]", faTokStr(fa.Tokenize("این،جمله")))

	// Arabic question: core keeps whole; FA splits
	require.Equal(t, "[س؟ت]", faTokStr(core.Tokenize("س؟ت")))
	require.Equal(t, "[س, ؟, ت]", faTokStr(fa.Tokenize("س؟ت")))

	// Arabic semicolon: core keeps whole; FA splits
	require.Equal(t, "[أ؛ب]", faTokStr(core.Tokenize("أ؛ب")))
	require.Equal(t, "[أ, ؛, ب]", faTokStr(fa.Tokenize("أ؛ب")))

	// ASCII hyphen: core and FA keep whole; AR splits (FA suffix lacks -)
	require.Equal(t, "[well-known]", faTokStr(core.Tokenize("well-known")))
	require.Equal(t, "[well-known]", faTokStr(fa.Tokenize("well-known")))
	require.Equal(t, "[well, -, known]", faTokStr(ar.Tokenize("well-known")))

	// same Arabic punctuation split for FA and AR
	require.Equal(t, faTokStr(fa.Tokenize("أ،ب؛ج؟")), faTokStr(ar.Tokenize("أ،ب؛ج؟")))
}
