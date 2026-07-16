package sk

// Twin of SlovakTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of SlovakTest.testLanguage
func TestSlovak_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sk")
	require.Equal(t, "sk", lt.GetLanguageCode())
	sents := lt.Analyze("Toto je testovací text.")
	require.NotEmpty(t, sents)
}
