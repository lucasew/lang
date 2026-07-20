package da

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("da")
	RegisterCoreDanishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "DA_I_I")
}

// Java Danish.getRelevantRules exact ID set.
func TestRegisterCoreDanishRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("da")
	RegisterCoreDanishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "SENTENCE_WHITESPACE", "WORD_REPEAT_RULE",
		"DA_UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}
