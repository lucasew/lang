package km

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Java Khmer.getRelevantRules exact ID set — no SharedLayout invent extras.
func TestRegisterCoreKhmerRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("km")
	RegisterCoreKhmerRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"HUNSPELL_RULE",
		"KM_SIMPLE_REPLACE",
		"KM_WORD_REPEAT_RULE",
		"KM_UNPAIRED_BRACKETS",
		"KM_SPACE_BEFORE_CONJUNCTION",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"COMMA_PARENTHESIS_WHITESPACE", "DOUBLE_PUNCTUATION", "UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START", "WHITESPACE_RULE", "SENTENCE_WHITESPACE",
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "WORD_REPEAT_RULE", "WORD_REPEAT_BEGINNING_RULE",
		"KM_WORD_REPEAT_BEGINNING_RULE",
	} {
		require.NotContains(t, ids, bad)
	}
}
