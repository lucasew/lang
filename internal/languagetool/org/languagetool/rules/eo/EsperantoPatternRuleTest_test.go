package eo

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEsperantoPatternRule_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("eo").Analyze("x"))
}
