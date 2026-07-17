package languagetool

// Twin of SL JLanguageToolTest — Check inject.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testSlovenian
func TestJLanguageTool_lang_sl_Slovenian(t *testing.T) {
	lt := NewJLanguageTool("sl")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "sl", lt.GetLanguageCode())
	require.Empty(t, lt.Check("To je preizkus."))
	require.NotEmpty(t, lt.Check("To je je preizkus."))
}
