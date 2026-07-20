package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("it")
	RegisterCoreItalianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "IT_A_IL")
}

// Java Italian.getRelevantRules exact ID set.
func TestRegisterCoreItalianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	RegisterCoreItalianRules(lt)
	require.ElementsMatch(t, language.ItalianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE",
		"IT_UNPAIRED_BRACKETS", "WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
