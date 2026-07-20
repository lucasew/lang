package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "RU_В_В")
}

// Java Russian.getRelevantRules exact ID set.
func TestRegisterCoreRussianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	require.ElementsMatch(t, language.RussianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"DOUBLE_PUNCTUATION", "EMPTY_LINE", "UNPAIRED_BRACKETS",
		"WHITESPACE_PUNCTUATION", "PUNCTUATION_PARAGRAPH_END",
		"RU_WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
