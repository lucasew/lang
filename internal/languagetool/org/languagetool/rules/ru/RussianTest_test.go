package ru

// Twin of RussianTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRussian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	require.Equal(t, "ru", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Это тест."))
}
