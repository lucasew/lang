package languagetool

// Twin of EsperantoTest — Analyze surface (full metadata in language package).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEsperanto_Test(t *testing.T) {
	lt := NewJLanguageTool("eo")
	require.Equal(t, "eo", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Saluton, mondo!"))
}
