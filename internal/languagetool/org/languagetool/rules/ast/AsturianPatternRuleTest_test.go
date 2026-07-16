package ast

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAsturianPatternRuleTest_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("ast").Analyze("x"))
}
