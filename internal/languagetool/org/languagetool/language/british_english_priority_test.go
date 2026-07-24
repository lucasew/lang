package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBritishEnglishPriorityForId(t *testing.T) {
	// Java BritishEnglish.id2prio
	require.Equal(t, -20, BritishEnglishPriorityForId("OXFORD_SPELLING_ISATION_NOUNS"))
	require.Equal(t, -21, BritishEnglishPriorityForId("OXFORD_SPELLING_ISE_VERBS"))
	require.Equal(t, -22, BritishEnglishPriorityForId("OXFORD_SPELLING_IZE"))
	// super English
	require.Equal(t, 1, BritishEnglishPriorityForId("I_A"))
	require.Equal(t, -1, BritishEnglishPriorityForId("EN_A_VS_AN"))
	// map size
	require.Equal(t, 3, len(BritishEnglishPriorityMap()))
}

func TestEnglishPriorityForIdForCode(t *testing.T) {
	gb := EnglishPriorityForIdForCode("en-GB")
	require.Equal(t, -20, gb("OXFORD_SPELLING_ISATION_NOUNS"))
	us := EnglishPriorityForIdForCode("en-US")
	// US does not use British map — Oxford ids fall through English/base (0 or base)
	require.Equal(t, 0, us("OXFORD_SPELLING_ISATION_NOUNS"))
	require.Equal(t, 1, us("I_A"))
	require.True(t, isBritishEnglishCode("en-GB"))
	require.True(t, isBritishEnglishCode("en_GB"))
	require.False(t, isBritishEnglishCode("en-US"))
	require.False(t, isBritishEnglishCode("en"))
}
