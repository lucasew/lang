package languagetool

// Twin of JA JLanguageToolTest — Check inject (script tokenize soft).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testJapanese
func TestJLanguageTool_lang_ja_Japanese(t *testing.T) {
	lt := NewJLanguageTool("ja")
	// word-repeat soft on plain tokens if any
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "ja", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("これはテストです。"))
	// check returns without panic
	_ = lt.Check("これはテストです。")
}
