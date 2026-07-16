package br

// Twin of BretonTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBreton_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("br")
	require.Equal(t, "br", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Un dest eo."))
}
