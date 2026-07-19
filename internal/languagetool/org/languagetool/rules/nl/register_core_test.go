package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("nl")
	RegisterCoreDutchRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "NL_ALS_OF")
}

func TestRegisterCoreDutchRules_RelevantRuleIDs(t *testing.T) {
	// Java Dutch.getRelevantRules getId surface (order free; presence check).
	lt := languagetool.NewJLanguageTool("nl")
	RegisterCoreDutchRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"MORFOLOGIK_RULE_NL_NL",
		"WHITESPACE_RULE",
		"NL_COMPOUNDS",
		"DUTCH_WRONG_WORD_IN_CONTEXT",
		"NL_WORD_COHERENCY",
		"NL_SIMPLE_REPLACE",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"NL_PREFERRED_WORD_RULE",
		"NL_SPACE_IN_COMPOUND",
		"SENTENCE_WHITESPACE",
		"NL_CHECKCASE",
	}
	for _, id := range want {
		require.Contains(t, ids, id, "missing Java getRelevantRules id %s", id)
	}
	// non-faithful invent IDs must not reappear
	require.NotContains(t, ids, "TOO_LONG_SENTENCE_NL")
	require.NotContains(t, ids, "NL_SENTENCE_WHITESPACE")
}
