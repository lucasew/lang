package sl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("sl")
	RegisterCoreSlovenianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "SL_V_V")
}

// Java Slovenian.getRelevantRules exact ID set.
func TestRegisterCoreSlovenianRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("sl")
	RegisterCoreSlovenianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_SL_SI",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WHITESPACE_PUNCTUATION", "SENTENCE_WHITESPACE", "SL_UNPAIRED_BRACKETS",
	} {
		require.NotContains(t, ids, bad)
	}
}
