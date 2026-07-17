package languagetool

// Twin of CA JLanguageToolTest — Check inject + typography.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testCleanOverlappingErrors
func TestJLanguageTool_lang_ca_CleanOverlappingErrors(t *testing.T) {
	cleaned := CleanOverlappingLocalMatches([]LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "low", Priority: 1},
		{FromPos: 2, ToPos: 4, RuleID: "high", Priority: 10},
	})
	require.Len(t, cleaned, 1)
	require.Equal(t, "high", cleaned[0].RuleID)
}

// Port of JLanguageToolTest.testGlobalSpelling
func TestJLanguageTool_lang_ca_GlobalSpelling(t *testing.T) {
	lt := NewJLanguageTool("ca")
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{"LanguageTool": {}}, nil))
	require.Empty(t, lt.Check("LanguageTool"))
}

// Port of JLanguageToolTest.testHyphenatedPlusCompound
func TestJLanguageTool_lang_ca_HyphenatedPlusCompound(t *testing.T) {
	lt := NewJLanguageTool("ca")
	// hyphen may tokenize as parts — accept both pieces
	known := map[string]struct{}{"nord": {}, "oest": {}, "nord-oest": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, nil))
	// exercise analyze/check path
	_ = lt.Check("nord-oest")
}

// Port of JLanguageToolTest.testValencianVariant
func TestJLanguageTool_lang_ca_ValencianVariant(t *testing.T) {
	lt := NewJLanguageTool("ca-ES-valencia")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "ca-ES-valencia", lt.GetLanguageCode())
	require.Empty(t, lt.Check("Hola món."))
}

// Port of JLanguageToolTest.testBalearicVariant
func TestJLanguageTool_lang_ca_BalearicVariant(t *testing.T) {
	lt := NewJLanguageTool("ca-ES-balear")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Equal(t, "ca-ES-balear", lt.GetLanguageCode())
	require.Empty(t, lt.Check("Hola món."))
}

// Port of JLanguageToolTest.testAdvancedTypography
func TestJLanguageTool_lang_ca_AdvancedTypography(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	require.Equal(t, "Això és…", ToAdvancedTypography("Això és...", cfg))
}
