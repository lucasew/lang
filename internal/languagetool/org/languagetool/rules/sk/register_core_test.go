package sk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("sk")
	RegisterCoreSlovakRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "SK_V_V")
}

// Java Slovak.getRelevantRules exact ID set and order.
func TestRegisterCoreSlovakRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sk")
	RegisterCoreSlovakRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SK_COMPOUNDS",
		"MORFOLOGIK_RULE_SK_SK",
	}
	// Order matches Java Arrays.asList / SlovakRelevantRuleIDs.
	require.Equal(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "SENTENCE_WHITESPACE", "SK_UNPAIRED_BRACKETS",
		"SK_VES_RULE", "SK_VES",
	} {
		require.NotContains(t, ids, bad)
	}
}
