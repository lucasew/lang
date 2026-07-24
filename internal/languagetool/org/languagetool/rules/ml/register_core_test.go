package ml

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java Malayalam.getRelevantRules exact ID set (layout subset + speller + word-repeat).
func TestRegisterCoreMalayalamRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ml")
	RegisterCoreMalayalamRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_ML_IN",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "SENTENCE_WHITESPACE", "WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN", "PUNCTUATION_PARAGRAPH_END",
	} {
		require.NotContains(t, ids, bad)
	}
}
