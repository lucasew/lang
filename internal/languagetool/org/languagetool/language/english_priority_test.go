package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishPriorityMap_Size(t *testing.T) {
	m := EnglishPriorityMap()
	require.Equal(t, 326, len(m))
	// Defensive copy
	m["I_A"] = 999
	require.Equal(t, 1, EnglishPriorityForId("I_A"))
}

func TestEnglishPriorityForId_MapSpotChecks(t *testing.T) {
	// Java id2prio spot-checks — not invent
	require.Equal(t, 1, EnglishPriorityForId("I_A"))
	require.Equal(t, -1, EnglishPriorityForId("EN_A_VS_AN"))
	require.Equal(t, -10, EnglishPriorityForId("MORFOLOGIK_RULE_EN_US"))
	require.Equal(t, 2, EnglishPriorityForId("ABBREVIATION_PUNCTUATION"))
	require.Equal(t, 3, EnglishPriorityForId("YOU_GOOD"))
	require.Equal(t, -4, EnglishPriorityForId("Y_ALL"))
	require.Equal(t, 1, EnglishPriorityForId("ACCESS_EXCESS"))
}

func TestEnglishPriorityForId_Prefixes(t *testing.T) {
	// Java getPriorityForId after map miss
	require.Equal(t, 2, EnglishPriorityForId("EN_COMPOUNDS_FOO"))
	require.Equal(t, -2, EnglishPriorityForId("PRP_VBZ"))
	require.Equal(t, -20, EnglishPriorityForId("CONFUSION_RULE_BAR"))
	require.Equal(t, -12, EnglishPriorityForId("EN_UPPER_CASE_NGRAM"))
	require.Equal(t, -9, EnglishPriorityForId("AI_SPELLING_RULE_X"))
	require.Equal(t, -9, EnglishPriorityForId("EN_MULTITOKEN_SPELLING_X"))
	require.Equal(t, -5, EnglishPriorityForId("EN_GB_SIMPLE_REPLACE_X"))
	require.Equal(t, -5, EnglishPriorityForId("EN_US_SIMPLE_REPLACE_X"))
	require.Equal(t, -51, EnglishPriorityForId("QB_EN_OXFORD"))
	require.Equal(t, -49, EnglishPriorityForId("EN_SIMPLE_REPLACE_PROGRAMME"))
	require.Equal(t, -49, EnglishPriorityForId("EN_SIMPLE_REPLACE_PROGRAMMES"))
	// AI_HYDRA_LEO branches
	require.Equal(t, -51, EnglishPriorityForId("AI_HYDRA_LEO_MISSING_COMMA"))
	require.Equal(t, -1, EnglishPriorityForId("AI_HYDRA_LEO_CP_YOU_YOUARE_X"))
	require.Equal(t, 2, EnglishPriorityForId("AI_HYDRA_LEO_CP_OTHER"))
	require.Equal(t, -14, EnglishPriorityForId("AI_HYDRA_LEO_MISSING_TO_X"))
	require.Equal(t, -11, EnglishPriorityForId("AI_HYDRA_LEO_OTHER"))
	require.Equal(t, -11, EnglishPriorityForId("AI_EN_LECTOR_X"))
	// FALSE_FRIENDS_PATTERN
	require.Equal(t, -21, EnglishPriorityForId("EN_FOR_DE_SPEAKERS_FALSE_FRIENDS_FOO"))
	// base Language
	require.Equal(t, -50, EnglishPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, EnglishPriorityForId("COMPLETELY_UNKNOWN_EN_XYZ"))
}
