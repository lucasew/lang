package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("es")
	require.Equal(t, "es", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Esto es un texto de prueba.`))
}
