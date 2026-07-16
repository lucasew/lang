package fa

// Twin of PersianTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of PersianTest.testLanguage
func TestPersian_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fa")
	require.Equal(t, "fa", lt.GetLanguageCode())
	sents := lt.Analyze("این یک متن آزمایشی است.")
	require.NotEmpty(t, sents)
}
