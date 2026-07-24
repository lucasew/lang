package sl

// Twin of SlovenianTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSlovenian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sl")
	require.Equal(t, "sl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("To je test."))
}
