package de

// Twin of GermanTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GermanTest.testLanguage
func TestGerman_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	require.Equal(t, "de", lt.GetLanguageCode())
	sents := lt.Analyze("Das ist ein Testtext.")
	require.NotEmpty(t, sents)
}
