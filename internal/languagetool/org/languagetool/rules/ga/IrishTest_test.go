package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIrish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ga")
	require.Equal(t, "ga", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Seo téacs tástála.`))
}
