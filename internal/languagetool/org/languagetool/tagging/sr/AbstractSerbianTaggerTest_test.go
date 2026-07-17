package sr

// Twin of AbstractSerbianTaggerTest (Java has no @Test) — MapWordTagger inject surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of AbstractSerbianTaggerTest (no @Test)
func TestAbstractSerbianTagger_NoTests(t *testing.T) {
	wt := tagging.MapWordTagger{
		"здраво": {tagging.NewTaggedWord("здраво", "interj")},
		"тест":   {tagging.NewTaggedWord("тест", "noun")},
	}
	ek := NewEkavianTagger(wt)
	require.Equal(t, EkavianDictionaryPath, ek.GetDictionaryPath())
	// Tag single word via BaseTagger if available
	if ek.BaseTagger != nil && ek.BaseTagger.WordTagger != nil {
		tags := ek.BaseTagger.WordTagger.Tag("здраво")
		require.NotEmpty(t, tags)
	}

	jk := NewJekavianTagger(wt)
	require.Equal(t, JekavianDictionaryPath, jk.GetDictionaryPath())

	sr := NewSerbianTagger(wt)
	require.Equal(t, EkavianDictionaryPath, sr.GetDictionaryPath())
}
