package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("ar")
	RegisterCoreArabicRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "AR_FI_FI")
}

// Java Arabic.getRelevantRules exact ID set.
func TestRegisterCoreArabicRules_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ar")
	RegisterCoreArabicRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	want := []string{
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"UNPAIRED_BRACKETS",
		"COMMA_PARENTHESIS_WHITESPACE",
		"TOO_LONG_SENTENCE",
		"HUNSPELL_RULE_AR",
		"ARABIC_COMMA_PARENTHESIS_WHITESPACE",
		"ARABIC_QM_WHITESPACE",
		"ARABIC_SC_WHITESPACE",
		"ARABIC_DOUBLE_PUNCTUATION",
		"ARABIC_WORD_REPEAT_RULE",
		"AR_SIMPLE_REPLACE",
		"AR_DIACRITICS_REPLACE",
		"AR_DARJA_REPLACE",
		"AR_HOMOPHONES_REPLACE",
		"AR_REDUNDANCY_REPLACE",
		"AR_WORD_COHERENCY",
		"AR_WORDINESS_REPLACE",
		"ARABIC_WRONG_WORD_IN_CONTEXT",
		"AR_VERB_TRANSITIVE_IINDIRECT",
		"AR_INFLECTED_ONE_WORD",
	}
	require.ElementsMatch(t, want, ids)
	for _, bad := range []string{
		"EMPTY_LINE", "TOO_LONG_PARAGRAPH", "UPPERCASE_SENTENCE_START",
		"WHITESPACE_PUNCTUATION", "AR_UNPAIRED_BRACKETS", "TOO_LONG_SENTENCE_AR",
		"AR_SENTENCE_WHITESPACE", "DOUBLE_PUNCTUATION",
	} {
		require.NotContains(t, ids, bad)
	}
}
