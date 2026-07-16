package da

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDanishPatternRule_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("da")
	require.NotEmpty(t, lt.Analyze("x"))
}
