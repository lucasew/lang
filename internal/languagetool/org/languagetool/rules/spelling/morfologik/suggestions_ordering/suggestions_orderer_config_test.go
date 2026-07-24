package suggestions_ordering

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionsOrdererConfig(t *testing.T) {
	t.Setenv(propMLSuggestionsOrdering, "")
	require.False(t, IsMLSuggestionsOrderingEnabled())
	SetMLSuggestionsOrderingEnabled(true)
	require.True(t, IsMLSuggestionsOrderingEnabled())
	SetMLSuggestionsOrderingEnabled(false)
	require.False(t, IsMLSuggestionsOrderingEnabled())

	SetNgramsPath("/tmp/ngrams")
	require.Equal(t, "/tmp/ngrams", GetNgramsPath())
}
