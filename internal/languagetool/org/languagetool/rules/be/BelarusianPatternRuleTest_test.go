package be

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestBelarusianPatternRule_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	require.NotEmpty(t, lt.Analyze("x"))
}
