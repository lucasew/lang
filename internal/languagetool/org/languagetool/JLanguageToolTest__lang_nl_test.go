package languagetool

// Twin of NL JLanguageToolTest — Check inject + typography.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testDutch
func TestJLanguageTool_lang_nl_Dutch(t *testing.T) {
	lt := NewJLanguageTool("nl")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("Dit is een zin zonder fouten."))
	require.NotEmpty(t, lt.Check("Dit is is een zin."))
}

// Port of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_nl_AdvancedTypography(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	require.Equal(t, "Dit is…", ToAdvancedTypography("Dit is...", cfg))
	require.Contains(t, ToAdvancedTypography("z. B.", cfg), "\u00a0")
}
