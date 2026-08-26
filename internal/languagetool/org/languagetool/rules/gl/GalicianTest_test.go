package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGalician_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("gl")
	require.Equal(t, "gl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Isto é un texto de proba.`))
}
