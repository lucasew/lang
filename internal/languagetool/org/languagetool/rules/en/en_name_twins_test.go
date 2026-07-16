package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamedConstructors(t *testing.T) {
	require.Equal(t, MorfologikAmericanSpellerRuleID, NewMorfologikAmericanSpellerRule().GetID())
	require.Equal(t, MorfologikBritishSpellerRuleID, NewMorfologikBritishSpellerRule().GetID())
	require.Equal(t, "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS", NewEnglishForGermansFalseFriendRule().GetID())
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", NewEnglishForFrenchFalseFriendRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", NewStyleTooOftenUsedNounRule().ID)
	require.Equal(t, "METRIC_UNITS_EN_US", NewUnitConversionRuleUS(nil).GetID())
	require.Equal(t, "METRIC_UNITS_EN_IMPERIAL", NewUnitConversionRuleImperial(nil).GetID())
}
