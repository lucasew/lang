package ro

// Twin of RomanianTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of RomanianTest.testLanguage
func TestRomanian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ro")
	require.Equal(t, "ro", lt.GetLanguageCode())
	sents := lt.Analyze("Acesta este un text de test.")
	require.NotEmpty(t, sents)
}
