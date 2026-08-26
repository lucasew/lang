package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/DemoTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DemoTest.testLanguage — JLanguageTool with demo short code (no language import cycle).
func TestDemo_Language(t *testing.T) {
	lt := NewJLanguageTool("xx")
	require.Equal(t, "xx", lt.GetLanguageCode())
	sents := lt.Analyze("This is a test.")
	require.NotEmpty(t, sents)
}
