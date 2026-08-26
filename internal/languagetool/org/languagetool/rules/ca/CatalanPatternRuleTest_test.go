package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanPatternRule_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	require.NotEmpty(t, lt.Analyze("x"))
}
