package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "FA_در_در")
}

// Java Persian.getRelevantRules exact ID set.
func TestRegisterCorePersianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"WHITESPACE_RULE",
		"TOO_LONG_SENTENCE",
		"PERSIAN_COMMA_PARENTHESIS_WHITESPACE",
		"PERSIAN_DOUBLE_PUNCTUATION",
		"PERSIAN_WORD_REPEAT_BEGINNING_RULE",
		"PERSIAN_WORD_REPEAT_RULE",
		"FA_SIMPLE_REPLACE",
		"FA_SPACE_BEFORE_CONJUNCTION",
		"FA_WORD_COHERENCY",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START", "SENTENCE_WHITESPACE", "WHITESPACE_PUNCTUATION",
		"TOO_LONG_SENTENCE_FA",
	} {
		require.NotContains(t, ids, bad)
	}
}
