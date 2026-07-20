package eo

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java Esperanto.getRelevantRules exact ID set.
func TestRegisterCoreEsperantoRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("eo")
	RegisterCoreEsperantoRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN", "PUNCTUATION_PARAGRAPH_END",
	} {
		require.NotContains(t, ids, bad)
	}
}
