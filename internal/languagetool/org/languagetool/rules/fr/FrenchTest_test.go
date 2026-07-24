package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFrench_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	require.Equal(t, "fr", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Bonjour le monde."))
}
