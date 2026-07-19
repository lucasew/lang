package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicMetadata(t *testing.T) {
	require.Equal(t, "ar", Arabic.GetShortCode())
	require.Equal(t, "Arabic", Arabic.GetName())
	require.Contains(t, Arabic.GetCountries(), "SA")
	require.Equal(t, "HUNSPELL_RULE_AR", Arabic.GetDefaultSpellingRuleID())
	require.Equal(t, "Taha Zerrouki", Arabic.GetMaintainers()[0].Name)
	require.Equal(t, "Sohaib Afifi", Arabic.GetMaintainers()[1].Name)
}

func TestArabicRelevantRuleIDs_MatchJavaGetId(t *testing.T) {
	// Faithful class getId / RULE_ID order from Arabic.getRelevantRules
	ids := ArabicRelevantRuleIDs()
	require.Equal(t, []string{
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
		"AR_VERB_TRANSITIVE_IINDIRECT", // upstream typo preserved
		"AR_INFLECTED_ONE_WORD",
	}, ids)
	require.Equal(t, ids, Arabic.GetRelevantRuleIDs())
	// invent IDs must not reappear
	require.NotContains(t, ids, "MULTIPLE_WHITESPACE")
	require.NotContains(t, ids, "COMMA_WHITESPACE")
	require.NotContains(t, ids, "LONG_SENTENCE")
	require.NotContains(t, ids, "ARABIC_SIMPLE_REPLACE")
}
