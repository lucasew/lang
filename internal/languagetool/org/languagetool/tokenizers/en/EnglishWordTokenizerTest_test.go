package en_test

import (
	"strings"
	"testing"

	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	en "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.en.EnglishWordTokenizerTest.
// Java uses EnglishTagger.INSTANCE (english.dict) for apostrophe/hyphen keep;
// Go uses the same path via EnsureDefaultEnglishTagger — no invent surface lists.

func tokStr(tokens []string) string {
	return "[" + strings.Join(tokens, ", ") + "]"
}

func TestEnglishWordTokenizer_Tokenize(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	// Java: EnglishTagger.INSTANCE always available; wire real english.dict isTagged.
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, en.IsTaggedEN, "IsTaggedEN must be wired from EnglishTagger/english.dict")

	w := en.NewEnglishWordTokenizer()
	tokens := w.Tokenize("This is\u00A0a test")
	require.Equal(t, 7, len(tokens))
	require.Equal(t, "[This,  , is, \u00A0, a,  , test]", tokStr(tokens))

	tokens2 := w.Tokenize("This\rbreaks")
	require.Equal(t, 3, len(tokens2))
	require.Equal(t, "[This, \r, breaks]", tokStr(tokens2))

	tokens3 := w.Tokenize("Now this is-really!-a test.")
	require.Equal(t, 11, len(tokens3))
	require.Equal(t, "[Now,  , this,  , is-really, !, -, a,  , test, .]", tokStr(tokens3))

	tokens4 := w.Tokenize("Now this is- really!- a test.")
	require.Equal(t, 15, len(tokens4))
	require.Equal(t, "[Now,  , this,  , is, -,  , really, !, -,  , a,  , test, .]", tokStr(tokens4))

	tokens5 := w.Tokenize("Now this is—really!—a test.")
	require.Equal(t, 13, len(tokens5))
	require.Equal(t, "[Now,  , this,  , is, —, really, !, —, a,  , test, .]", tokStr(tokens5))

	tokens6 := w.Tokenize("fo'c'sle")
	require.Equal(t, 1, len(tokens6))

	tokens7 := w.Tokenize("I'm John.")
	require.Equal(t, "[I, 'm,  , John, .]", tokStr(tokens7))
	require.Equal(t, 5, len(tokens7))

	tokens8 := w.Tokenize("You hadn’t.")
	require.Equal(t, "[You,  , had, n’t, .]", tokStr(tokens8))
	require.Equal(t, 5, len(tokens8))

	tokens9 := w.Tokenize("We'are")
	require.Equal(t, "[We, ', are]", tokStr(tokens9))
	require.Equal(t, 3, len(tokens9))

	tokens10 := w.Tokenize("'We're'")
	require.Equal(t, "[', We, 're, ']", tokStr(tokens10))
	require.Equal(t, 4, len(tokens10))

	tokens11 := w.Tokenize("'We’re the best.'")
	require.Equal(t, "[', We, ’re,  , the,  , best, ., ']", tokStr(tokens11))
	require.Equal(t, 9, len(tokens11))

	tokens12 := w.Tokenize("'Don't do it'")
	require.Equal(t, "[', Do, n't,  , do,  , it, ']", tokStr(tokens12))
	require.Equal(t, 8, len(tokens12))

	tokens13 := w.Tokenize("‘Don’t do it’")
	require.Equal(t, "[‘, Do, n’t,  , do,  , it, ’]", tokStr(tokens13))
	require.Equal(t, 8, len(tokens13))

	tokens14 := w.Tokenize("Don't do it")
	require.Equal(t, "[Do, n't,  , do,  , it]", tokStr(tokens14))
	require.Equal(t, 6, len(tokens14))

	tokens15 := w.Tokenize("My address is address@email.com")
	require.Equal(t, "[My,  , address,  , is,  , address@email.com]", tokStr(tokens15))
	require.Equal(t, 7, len(tokens15))

	tokens16 := w.Tokenize("@test@test.social you are aweesome!")
	require.Equal(t, "[@test@test.social,  , you,  , are,  , aweesome, !]", tokStr(tokens16))
	require.Equal(t, 8, len(tokens16))

	tokens17 := w.Tokenize("My address is address@email.com or other@email.com.")
	require.Equal(t, "[My,  , address,  , is,  , address@email.com,  , or,  , other@email.com, .]", tokStr(tokens17))
	require.Equal(t, 12, len(tokens17))

	tokens18 := w.Tokenize("doin' that")
	require.Equal(t, "[doin',  , that]", tokStr(tokens18))
	require.Equal(t, 3, len(tokens18))

	tokens19 := w.Tokenize("ne’er e'er o’er jack-o'-lantern")
	require.Equal(t, "[ne’er,  , e'er,  , o’er,  , jack-o'-lantern]", tokStr(tokens19))
	require.Equal(t, 7, len(tokens19))

	tokens20 := w.Tokenize("I'm a cool test\u000Bwith a line")
	require.Equal(t, "[I, 'm,  , a,  , cool,  , test, \u000B, with,  , a,  , line]", tokStr(tokens20))
	require.Equal(t, 14, len(tokens20))

	tokens21 := w.Tokenize("fast⇿superfast")
	require.Equal(t, "[fast, ⇿, superfast]", tokStr(tokens21))
	require.Equal(t, 3, len(tokens21))
}
