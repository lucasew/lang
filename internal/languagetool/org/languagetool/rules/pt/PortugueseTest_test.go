package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortuguese_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt")
	require.Equal(t, "pt", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Isto é um texto de teste.`))
}
