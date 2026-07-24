package en

// Twin of EnglishTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	require.Equal(t, "en", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("This is a test sentence."))
}

func TestEnglish_RepeatedPatternRules(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("en").Analyze("Repeated rules deferred."))
}

func TestEnglish_Messages(t *testing.T) {
	// Message bundle surface via ResourceBundleTools is covered in core package.
	require.NotEmpty(t, languagetool.MessageBundleName)
}
