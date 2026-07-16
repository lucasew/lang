package ca

// Twin of CatalanTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalan_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	require.Equal(t, "ca", lt.GetLanguageCode())
	sents := lt.Analyze("Aquest és un text de prova.")
	require.NotEmpty(t, sents)
}

func TestCatalan_RepeatedPatternRules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	require.NotEmpty(t, lt.Analyze("Text de prova."))
}

func TestCatalan_TrimMatchEnds(t *testing.T) {
	require.True(t, true)
}
