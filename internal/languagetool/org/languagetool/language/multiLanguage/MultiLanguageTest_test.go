package multiLanguage

// Twin of MultiLanguageTest (Java @Ignore FastText) — identifier inject smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

// Port of MultiLanguageTest (no active @Test without FastText)
func TestMultiLanguage_NoTests(t *testing.T) {
	// soft multi-lang: SimpleLanguageIdentifier with inject spellers
	spellers := map[string]identifier.SpellerFunc{
		"en": func(w string) bool {
			switch w {
			case "hello", "world", "This", "is", "an", "English", "test":
				return false
			default:
				return true
			}
		},
		"de": func(w string) bool {
			switch w {
			case "Hier", "kommt", "ein", "deutscher", "Satz":
				return false
			default:
				return true
			}
		},
	}
	id := identifier.NewSimpleLanguageIdentifierWith([]string{"en", "de"}, spellers)
	// mixed text detection surface
	d := id.Detect("This is an English test", nil, []string{"en", "de"})
	if d != nil {
		require.NotEmpty(t, d.GetDetectedLanguageCode())
	}
	// annotate path
	a := languagetool.NewLanguageAnnotator()
	frags := a.DetectLanguages("This is English. Hier kommt Deutsch.", "en", []string{"de"})
	require.NotEmpty(t, frags)
}
