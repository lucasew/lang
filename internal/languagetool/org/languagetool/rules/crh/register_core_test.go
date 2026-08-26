package crh

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java CrimeanTatar.getRelevantRules exact ID set.
func TestRegisterCoreCrimeanTatarRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("crh")
	RegisterCoreCrimeanTatarRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"MORFOLOGIK_RULE_CRH_UA",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"UNPAIRED_BRACKETS", "EMPTY_LINE", "TOO_LONG_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE", "WHITESPACE_PUNCTUATION",
		"WORD_REPEAT_RULE", "PUNCTUATION_PARAGRAPH_END",
	} {
		require.NotContains(t, ids, bad)
	}
}
