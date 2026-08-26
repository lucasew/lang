package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageModelRelevantRuleIDs(t *testing.T) {
	de := GermanLanguageModelRelevantRuleIDs()
	require.Equal(t, []string{"DE_UPPER_CASE_NGRAM", "CONFUSION_RULE", "DE_PROHIBITED_COMPOUNDS"}, de)
	require.Equal(t, []string{"GERMAN_SPELLER_RULE"}, GermanyGermanLanguageModelCapableRuleIDs())

	en := EnglishLanguageModelRelevantRuleIDs()
	require.Equal(t, []string{"EN_UPPER_CASE_NGRAM", "CONFUSION_RULE", "NGRAM_RULE"}, en)
	require.Equal(t, []string{"EN_FOR_DE_SPEAKERS_FALSE_FRIENDS"}, EnglishLanguageModelCapableRuleIDsForMotherTongue("de"))
	require.Equal(t, []string{"EN_FOR_FR_SPEAKERS_FALSE_FRIENDS"}, EnglishLanguageModelCapableRuleIDsForMotherTongue("fr"))
	require.Equal(t, []string{"EN_FOR_ES_SPEAKERS_FALSE_FRIENDS"}, EnglishLanguageModelCapableRuleIDsForMotherTongue("es"))
	require.Equal(t, []string{"EN_FOR_NL_SPEAKERS_FALSE_FRIENDS"}, EnglishLanguageModelCapableRuleIDsForMotherTongue("nl"))
	require.Nil(t, EnglishLanguageModelCapableRuleIDsForMotherTongue(""))
	require.Nil(t, EnglishLanguageModelCapableRuleIDsForMotherTongue("it"))

	require.Equal(t, []string{"CONFUSION_RULE"}, FrenchLanguageModelRelevantRuleIDs())
	require.Equal(t, []string{"CONFUSION_RULE"}, PortugueseLanguageModelRelevantRuleIDs())
	require.Equal(t, []string{"CONFUSION_RULE"}, SpanishLanguageModelRelevantRuleIDs())
	require.Equal(t, []string{"CONFUSION_RULE"}, RussianLanguageModelRelevantRuleIDs())
}
