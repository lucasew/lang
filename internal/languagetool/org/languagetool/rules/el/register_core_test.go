package el

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("el")
	RegisterCoreGreekRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "EL_ΚΑΙ_ΚΑΙ")
}

// Java Greek.getRelevantRules exact ID set.
func TestRegisterCoreGreekRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("el")
	RegisterCoreGreekRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"EL_UNPAIRED_BRACKETS",
		"TOO_LONG_SENTENCE",
		"MORFOLOGIK_RULE_EL_GR",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"GREEK_WORD_REPEAT_BEGINNING_RULE",
		"WORD_REPEAT_RULE",
		"GREEK_HOMONYMS_REPLACE",
		"EL_SPECIFIC_CASE",
		"GREEK_ORTHOGRAPHY_NUMERAL_STRESS",
		"EL_REDUNDANCY_REPLACE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "SENTENCE_WHITESPACE",
		"WHITESPACE_PUNCTUATION", "TOO_LONG_SENTENCE_EL", "UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}
