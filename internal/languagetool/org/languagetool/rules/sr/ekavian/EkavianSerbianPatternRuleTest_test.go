package ekavian

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEkavianSerbianPatternRule_Rules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("sr").Analyze("x"))
}
