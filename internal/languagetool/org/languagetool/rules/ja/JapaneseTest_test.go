package ja

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestJapanese_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ja")
	require.Equal(t, "ja", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`これはテストです。`))
}
