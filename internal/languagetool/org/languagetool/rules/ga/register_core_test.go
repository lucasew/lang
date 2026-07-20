package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("ga")
	RegisterCoreIrishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "GA_AGUS_AGUS")
}

// Java Irish.getRelevantRules exact ID set.
func TestRegisterCoreIrishRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ga")
	RegisterCoreIrishRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"UNPAIRED_BRACKETS",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_GA_IE",
		"GA_LOGAINM",
		"GA_PEOPLE",
		"GA_SPASANNA",
		"GA_COMPOUNDS",
		"GA_PRESTANDARD_REPLACE",
		"GA_REPLACE",
		"GA_FGB_EQ_REPLACE",
		"GA_ENGLISH_HOMOPHONE",
		"GA_DHA_NO_BEIRT",
		"GA_DATIVE_PLURALS_STD",
		"GA_SPECIFIC_CASE",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "WHITESPACE_PUNCTUATION", "GA_UNPAIRED_BRACKETS",
		"GA_SENTENCE_WHITESPACE", "TOO_LONG_SENTENCE_GA", "PUNCTUATION_PARAGRAPH_END",
	} {
		require.NotContains(t, ids, bad)
	}
}
