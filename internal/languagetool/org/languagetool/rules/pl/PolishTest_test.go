package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPolish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pl")
	require.Equal(t, "pl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`To jest zdanie testowe.`))
}
