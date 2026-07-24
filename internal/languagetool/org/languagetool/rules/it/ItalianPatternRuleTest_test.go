package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestItalianPatternRule_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	require.NotEmpty(t, lt.Analyze("x"))
}
