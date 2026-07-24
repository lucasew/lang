package ast

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAsturian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ast")
	require.Equal(t, "ast", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Esto ye un test.`))
}
