package be

// Twin of BelarusianTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of BelarusianTest.testLanguage
func TestBelarusian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	require.Equal(t, "be", lt.GetLanguageCode())
	sents := lt.Analyze("Гэта тэставы тэкст.")
	require.NotEmpty(t, sents)
}
