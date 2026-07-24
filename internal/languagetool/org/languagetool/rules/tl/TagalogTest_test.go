package tl

// Twin of TagalogTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of TagalogTest.testLanguage
func TestTagalog_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("tl")
	require.Equal(t, "tl", lt.GetLanguageCode())
	sents := lt.Analyze("Ito ay isang test.")
	require.NotEmpty(t, sents)
}
