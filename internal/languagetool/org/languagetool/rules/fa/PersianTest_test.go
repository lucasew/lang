package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPersian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fa")
	require.Equal(t, "fa", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`این یک آزمایش است.`))
}
