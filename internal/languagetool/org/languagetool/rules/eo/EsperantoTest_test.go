package eo

// Twin of EsperantoTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEsperanto_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("eo")
	require.Equal(t, "eo", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Saluton mondo."))
}
