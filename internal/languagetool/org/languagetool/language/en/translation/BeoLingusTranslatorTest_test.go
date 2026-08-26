package translation

// Twin of BeoLingusTranslatorTest — split/clean helpers; full dict deferred.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeoLingusTranslator_Split(t *testing.T) {
	require.Equal(t, []string{"foo"}, Split("foo"))
	require.Equal(t, []string{"foo { bar } foo"}, Split("foo { bar } foo"))
	require.Equal(t, []string{"foo { bar }", "foo"}, Split("foo { bar }; foo"))
	require.Equal(t, []string{"foo", "bar", "foo"}, Split("foo; bar; foo"))
	require.Equal(t, []string{"foo", "bar { blah }", "foo"}, Split("foo; bar { blah }; foo"))
	require.Equal(t, []string{"foo", "bar { blah; blubb }", "foo"}, Split("foo; bar { blah; blubb }; foo"))
	require.Equal(t, []string{"foo", "bar { blah; blubb; three four }", "foo"}, Split("foo; bar { blah; blubb; three four }; foo"))
}

func TestBeoLingusTranslator_CleanTranslationForReplace(t *testing.T) {
	require.Equal(t, "", CleanTranslationForReplace("", ""))
	require.Equal(t, "go", CleanTranslationForReplace("to go", ""))
	require.Equal(t, "to go", CleanTranslationForReplace("to go", "need"))
	require.Equal(t, "go", CleanTranslationForReplace("to go", "will"))
	require.Equal(t, "go", CleanTranslationForReplace("to go", "foo"))
	require.Equal(t, "go", CleanTranslationForReplace("to go", "to"))
	require.Equal(t, "foo", CleanTranslationForReplace("foo (bar) {mus}", ""))
	require.Equal(t, "some thing , something", CleanTranslationForReplace("some thing [Br.], something", ""))
	require.Equal(t, "Friday", CleanTranslationForReplace("Friday /Fri/", ""))
	require.Equal(t, "demise", CleanTranslationForReplace("demise [poet.] <death>", ""))
}

func TestBeoLingusTranslator_GetTranslationSuffix(t *testing.T) {
	require.Equal(t, "", GetTranslationSuffix(""))
	require.Equal(t, "", GetTranslationSuffix(" "))
	require.Equal(t, "", GetTranslationSuffix("foo bar"))
	require.Equal(t, "[Br.]", GetTranslationSuffix("foo bar [Br.]"))
	require.Equal(t, "{ugs} [Br.]", GetTranslationSuffix("foo bar {ugs} [Br.]"))
	require.Equal(t, "{ugs} [Br.] (Blah)", GetTranslationSuffix("foo bar {ugs} [Br.] (Blah)"))
	require.Equal(t, "<blah>", GetTranslationSuffix("foo bar <blah>"))
}

// Port of BeoLingusTranslatorTest.translateInflectedForm — inject dict (full Beolingus deferred).
func TestBeoLingusTranslator_TranslateInflectedForm(t *testing.T) {
	tr := NewBeoLingusTranslator()
	tr.Dict["gehen"] = []string{"to go"}
	tr.Inflected["ging"] = "gehen"
	// without German tagger: surface maps via Inflected inject
	got := tr.TranslateInflectedForm("ging", "will")
	require.Equal(t, []string{"go"}, got)
}

// Port of BeoLingusTranslatorTest.translate — inject dict.
func TestBeoLingusTranslator_Translate(t *testing.T) {
	tr := NewBeoLingusTranslator()
	tr.Dict["Haus"] = []string{"house; home"}
	tr.Dict["Handy"] = []string{"mobile phone [Br.]; cell phone [Am.]"}
	require.Equal(t, []string{"house", "home"}, tr.Translate("Haus", ""))
	got := tr.Translate("Handy", "")
	require.Contains(t, got, "mobile phone")
	require.Contains(t, got, "cell phone")
	require.Empty(t, tr.Translate("xyzzy", ""))
}

// Port of BeoLingusTranslatorTest.americanBritishVariants — soft AE/BE map.
func TestBeoLingusTranslator_AmericanBritishVariants(t *testing.T) {
	require.Equal(t, "colour", AmericanToBritish("color"))
	require.Equal(t, "Colour", AmericanToBritish("Color"))
	require.Equal(t, "centre", AmericanToBritish("center"))
	require.Equal(t, "house", AmericanToBritish("house")) // unchanged
	tr := NewBeoLingusTranslator()
	tr.Dict["Farbe"] = []string{"color"}
	got := tr.Translate("Farbe", "")
	require.Equal(t, []string{"color"}, got)
	require.Equal(t, "colour", AmericanToBritish(got[0]))
}
