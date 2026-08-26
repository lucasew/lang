package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDutch_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("nl")
	require.Equal(t, "nl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Dit is een testzin."))
}
