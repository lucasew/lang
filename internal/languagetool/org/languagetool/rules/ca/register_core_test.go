package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("ca")
	RegisterCoreCatalanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "CA_A_EL")
}

// Java Catalan.getRelevantRules exact ID set (non-Valencian).
func TestRegisterCoreCatalanRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	RegisterCoreCatalanRules(lt)
	require.ElementsMatch(t, language.CatalanRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "SENTENCE_WHITESPACE", "WHITESPACE_PUNCTUATION",
		"TOO_LONG_PARAGRAPH", "WHITESPACE_PARAGRAPH", "CA_WORD_COHERENCY_VALENCIA",
		// CatalanRepeatedWordsRule commented out in Java
		"CA_REPEATEDWORDS",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}

// Java ValencianCatalan adds WordCoherencyValencianRule.
func TestRegisterCoreCatalanRules_ValencianExtra(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca-ES-valencia")
	RegisterCoreCatalanRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "CA_WORD_COHERENCY_VALENCIA")
	// base Catalan set still present
	require.Contains(t, ids, "CATALAN_WORD_REPEAT_RULE")
	require.Contains(t, ids, "CA_SPLIT_LONG_SENTENCE")
}
