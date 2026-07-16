package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGalicianPatternRuleTest_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("gl").Analyze("x"))
}
