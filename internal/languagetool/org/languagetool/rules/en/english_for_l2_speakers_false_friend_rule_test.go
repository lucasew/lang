package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishForL2SpeakersFalseFriendRules(t *testing.T) {
	de := NewEnglishForGermansFalseFriendRule()
	require.Equal(t, "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS", de.GetID())
	require.Equal(t, []string{"confusion_sets_l2_de.txt"}, de.GetFilenames())
	require.Equal(t, "de", de.MotherTongue)

	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", NewEnglishForFrenchFalseFriendRule().GetID())
	require.Equal(t, "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS", NewEnglishForSpaniardsFalseFriendRule().GetID())
	require.Equal(t, "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS", NewEnglishForDutchmenFalseFriendRule().GetID())
}
