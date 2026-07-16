package sr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSerbianPatternRule_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("sr").Analyze("x"))
}
