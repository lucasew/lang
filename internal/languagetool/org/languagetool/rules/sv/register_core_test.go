package sv

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("sv")
	RegisterCoreSwedishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "SV_I_I")
}

// Java Swedish.getRelevantRules exact ID set.
func TestRegisterCoreSwedishRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sv")
	RegisterCoreSwedishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"TOO_LONG_PARAGRAPH",
		"UPPERCASE_SENTENCE_START",
		"TOO_LONG_SENTENCE",
		"WORD_REPEAT_RULE",
		"SV_WORD_COHERENCY",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"SV_COMPOUNDS",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"SV_UNPAIRED_BRACKETS", "TOO_LONG_SENTENCE_SV", "WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}
