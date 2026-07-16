package ta

// Twin of TamilTest.testLanguage — analyze smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTamil_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ta")
	require.Equal(t, "ta", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("இது ஒரு சோதனை."))
}
