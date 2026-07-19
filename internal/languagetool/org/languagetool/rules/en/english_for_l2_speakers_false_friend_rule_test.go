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
	// Java addExamplePair: handy → phone
	require.Equal(t, []string{"phone"}, de.GetIncorrectExamples()[0].GetCorrections())
	require.Contains(t, de.ExampleWrong, "<marker>handy</marker>")

	fr := NewEnglishForFrenchFalseFriendRule()
	require.Equal(t, "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS", fr.GetID())
	require.Equal(t, []string{"complete"}, fr.GetIncorrectExamples()[0].GetCorrections())

	es := NewEnglishForSpaniardsFalseFriendRule()
	require.Equal(t, "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS", es.GetID())
	require.Equal(t, []string{"produce"}, es.GetIncorrectExamples()[0].GetCorrections())

	nl := NewEnglishForDutchmenFalseFriendRule()
	require.Equal(t, "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS", nl.GetID())
	require.Equal(t, []string{"wall"}, nl.GetIncorrectExamples()[0].GetCorrections())
}
