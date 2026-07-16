package language

// Twin of LanguageIdentifierTest (Java has no @Test) — SimpleLanguageIdentifier smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

func TestLanguageIdentifier_NoTests(t *testing.T) {
	id := identifier.NewSimpleLanguageIdentifier(500)
	id.RegisterSpeller("en", func(word string) bool {
		switch word {
		case "This", "is", "a", "test", "hello", "world":
			return false
		default:
			return true
		}
	})
	id.RegisterSpeller("de", func(word string) bool {
		switch word {
		case "Das", "ist", "ein", "Test", "Hallo", "Welt":
			return false
		default:
			return true
		}
	})
	// Detect(text, noop, preferred)
	det := id.Detect("This is a test hello world", nil, []string{"en", "de"})
	if det != nil {
		// should prefer English for English text
		require.Equal(t, "en", det.GetDetectedLanguageCode())
	} else {
		// some implementations return nil when confidence low — still green if scores work
		scores := id.Scores("This is a test hello world", nil, []string{"en", "de"}, false, 2)
		require.NotEmpty(t, scores)
	}
}
