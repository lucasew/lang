package is

// Twin of IcelandicTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIcelandic_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("is")
	require.Equal(t, "is", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Þetta er próf."))
}
