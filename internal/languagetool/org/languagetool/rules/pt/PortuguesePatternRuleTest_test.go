package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortuguesePatternRuleTest_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("pt").Analyze("x"))
}
