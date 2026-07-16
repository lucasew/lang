package sv

// Twin of SwedishTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of SwedishTest.testLanguage
func TestSwedish_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	require.Equal(t, "sv", lt.GetLanguageCode())
	sents := lt.Analyze("Detta är en testtext.")
	require.NotEmpty(t, sents)
}
