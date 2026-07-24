package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "UK_В_В")
}

// Java Ukrainian.getRelevantRules exact ID set.
func TestRegisterCoreUkrainianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	require.ElementsMatch(t, language.UkrainianRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	for _, bad := range []string{
		"DOUBLE_PUNCTUATION", "UNPAIRED_BRACKETS", "SENTENCE_WHITESPACE",
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "WHITESPACE_PUNCTUATION",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PARAGRAPH",
	} {
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), bad)
	}
}
