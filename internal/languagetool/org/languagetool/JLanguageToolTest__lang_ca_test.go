package languagetool

// Twin of languagetool-language-modules/ca/src/test/java/org/languagetool/JLanguageToolTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testCleanOverlappingErrors
func TestJLanguageTool_lang_ca_CleanOverlappingErrors(t *testing.T) {
	lt := NewJLanguageTool("ca")
	require.Equal(t, "ca", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Això és una prova."))
}

// Port of JLanguageToolTest.testGlobalSpelling
func TestJLanguageTool_lang_ca_GlobalSpelling(t *testing.T) {
	lt := NewJLanguageTool("ca")
	require.NotEmpty(t, lt.Analyze("LanguageTool"))
}

// Port of JLanguageToolTest.testHyphenatedPlusCompound
func TestJLanguageTool_lang_ca_HyphenatedPlusCompound(t *testing.T) {
	lt := NewJLanguageTool("ca")
	require.NotEmpty(t, lt.Analyze("nord-oest"))
}

// Port of JLanguageToolTest.testValencianVariant
func TestJLanguageTool_lang_ca_ValencianVariant(t *testing.T) {
	lt := NewJLanguageTool("ca-ES-valencia")
	require.Equal(t, "ca-ES-valencia", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Hola món."))
}

// Port of JLanguageToolTest.testBalearicVariant
func TestJLanguageTool_lang_ca_BalearicVariant(t *testing.T) {
	lt := NewJLanguageTool("ca-ES-balear")
	require.Equal(t, "ca-ES-balear", lt.GetLanguageCode())
	require.NotEmpty(t, lt.Analyze("Hola món."))
}

// Port of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_ca_AdvancedTypography(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	require.Equal(t, "Això és…", ToAdvancedTypography("Això és...", cfg))
}
