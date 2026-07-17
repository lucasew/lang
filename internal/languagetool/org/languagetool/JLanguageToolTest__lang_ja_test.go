package languagetool

// Twin of languagetool-language-modules/ja/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testJapanese
func TestJLanguageTool_lang_ja_Japanese(t *testing.T) {
	lt := NewJLanguageTool("ja")
	require.Equal(t, "ja", lt.GetLanguageCode())
	// AnalyzePlain path (SRX may keep as one unit)
	require.NotEmpty(t, lt.Analyze("これはテストです。"))
}
