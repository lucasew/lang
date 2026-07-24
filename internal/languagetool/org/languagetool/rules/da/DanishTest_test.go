package da

// Twin of DanishTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of DanishTest.testLanguage
func TestDanish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("da")
	require.Equal(t, "da", lt.GetLanguageCode())
	sents := lt.Analyze("Dette er en testtekst.")
	require.NotEmpty(t, sents)
}
