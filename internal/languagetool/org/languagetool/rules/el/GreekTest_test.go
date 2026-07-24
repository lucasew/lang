package el

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGreek_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("el")
	require.Equal(t, "el", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze(`Αυτό είναι ένα κείμενο.`))
}
