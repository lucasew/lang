package sv

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSwedishPatternRule_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("sv").Analyze("x"))
}
