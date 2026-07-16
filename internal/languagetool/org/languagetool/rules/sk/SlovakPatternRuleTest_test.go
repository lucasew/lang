package sk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSlovakPatternRule_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sk")
	require.NotEmpty(t, lt.Analyze("x"))
}
