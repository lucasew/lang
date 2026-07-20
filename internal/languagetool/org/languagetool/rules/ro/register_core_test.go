package ro

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("ro")
	RegisterCoreRomanianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "RO_DE_DE")
}

// Java Romanian.getRelevantRules exact ID set.
func TestRegisterCoreRomanianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ro")
	RegisterCoreRomanianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"UNPAIRED_BRACKETS",
		"WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_RO_RO",
		"ROMANIAN_WORD_REPEAT_BEGINNING_RULE",
		"RO_SIMPLE_REPLACE",
		"RO_COMPOUND",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
	} {
		require.NotContains(t, ids, bad)
	}
}
