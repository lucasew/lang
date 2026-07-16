package languagetool

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testPolish — analysis smoke (full grammar deferred)
func TestJLanguageTool_lang_pl_Polish(t *testing.T) {
	lt := NewJLanguageTool("pl")
	require.Equal(t, "pl", lt.GetLanguageCode())
	sents := lt.Analyze("To jest zdanie. A to drugie.")
	require.NotEmpty(t, sents)
}
