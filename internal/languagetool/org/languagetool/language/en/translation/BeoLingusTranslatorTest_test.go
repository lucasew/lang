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

func TestBeoLingusTranslator_TranslateInflectedForm(t *testing.T) {
	t.Skip("unimplemented: full Beolingus dictionary + German tagger")
}
func TestBeoLingusTranslator_Translate(t *testing.T) {
	t.Skip("unimplemented: full Beolingus dictionary")
}
func TestBeoLingusTranslator_AmericanBritishVariants(t *testing.T) {
	t.Skip("unimplemented: full Beolingus dictionary")
}
