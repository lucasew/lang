package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("ru").Analyze("test"))
}
