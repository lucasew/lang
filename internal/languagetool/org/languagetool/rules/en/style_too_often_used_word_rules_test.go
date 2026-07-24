package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedWordRules_IDs(t *testing.T) {
	require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", NewStyleTooOftenUsedNounRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_VERB_EN", NewStyleTooOftenUsedVerbRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_EN", NewStyleTooOftenUsedAdjectiveRule().GetID())
}
