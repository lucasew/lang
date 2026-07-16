package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishPatternRuleTest_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("es").Analyze("x"))
}
