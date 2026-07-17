package languagetool

// Twin of languagetool-language-modules/sl/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testSlovenian
func TestJLanguageTool_lang_sl_Slovenian(t *testing.T) {
	lt := NewJLanguageTool("sl")
	require.Equal(t, "sl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("To je preizkus."))
}
