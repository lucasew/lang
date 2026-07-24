package it

// Twin of LanguageSpecificSpellcheckerTest — analyze/speller surface smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificSpellcheckerTest.testRules
func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	require.NotEmpty(t, lt.Analyze("test"))
}
