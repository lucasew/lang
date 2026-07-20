package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("fr")
	RegisterCoreFrenchRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "FR_MALGRE_QUE")
}

// Java French.getRelevantRules exact ID set.
func TestRegisterCoreFrenchRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	RegisterCoreFrenchRules(lt)
	require.ElementsMatch(t, language.FrenchRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "WHITESPACE_PUNCTUATION", "WHITESPACE_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WORD_REPEAT_RULE",
		"FR_SENTENCE_WHITESPACE", "TOO_LONG_SENTENCE_FR",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
