package languagetool

// Twin of ES JLanguageToolTest — Check inject.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testXMLRules
func TestJLanguageTool_lang_es_XMLRules(t *testing.T) {
	lt := NewJLanguageTool("es")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("Esto es una prueba."))
	require.NotEmpty(t, lt.Check("Esto es es una prueba."))
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_es_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("es")
	known := map[string]struct{}{"Buenos": {}, "Aires": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, nil))
	require.Empty(t, lt.Check("Buenos Aires"))
}

// Port of JLanguageToolTest.testFilterRuleMatches
func TestJLanguageTool_lang_es_FilterRuleMatches(t *testing.T) {
	lt := NewJLanguageTool("es")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{"El": {}, "gato": {}, "duerme": {}}, nil))
	// clean non-overlapping synthetic via CleanOverlappingLocalMatches
	raw := []LocalMatch{
		{FromPos: 0, ToPos: 2, RuleID: "A", Priority: 1},
		{FromPos: 1, ToPos: 5, RuleID: "B", Priority: 5},
	}
	cleaned := CleanOverlappingLocalMatches(raw)
	require.Len(t, cleaned, 1)
	require.Equal(t, "B", cleaned[0].RuleID)
	require.Empty(t, lt.Check("El gato duerme."))
}
