package languagetool

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testDutch
func TestJLanguageTool_lang_nl_Dutch(t *testing.T) {
	lt := NewJLanguageTool("nl")
	require.Equal(t, "nl", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Dit is een zin zonder fouten."))
}

// Port of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_nl_AdvancedTypography(t *testing.T) {
	// NL uses common advanced typography (ellipsis / nbsp abbreviations)
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	require.Equal(t, "Dit is…", ToAdvancedTypography("Dit is...", cfg))
	require.Contains(t, ToAdvancedTypography("z. B.", cfg), "\u00a0")
}
