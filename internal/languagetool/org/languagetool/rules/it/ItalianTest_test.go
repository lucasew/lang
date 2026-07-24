package it

// Twin of ItalianTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of ItalianTest.testLanguage
func TestItalian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	require.Equal(t, "it", lt.GetLanguageCode())
	sents := lt.Analyze("Inserite qui il vostro testo.")
	require.NotEmpty(t, sents)
}
