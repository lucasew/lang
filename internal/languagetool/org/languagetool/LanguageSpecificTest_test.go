package languagetool

// Twin of LanguageSpecificTest (Java has no @Test) — language short-code Analyze smoke.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificTest (no @Test)
func TestLanguageSpecific_NoTests(t *testing.T) {
	for _, code := range []string{"en", "de", "fr", "es", "pl", "uk", "nl", "pt"} {
		lt := NewJLanguageTool(code)
		require.Equal(t, code, lt.GetLanguageCode())
		require.NotEmpty(t, lt.Analyze("test"), code)
	}
}
