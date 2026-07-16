package suggestions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionsChanges(t *testing.T) {
	t.Cleanup(ResetSuggestionsChanges)
	require.False(t, IsRunningExperiment("x"))

	s := InitSuggestionsChanges(&SuggestionChangesTestConfig{Language: "en"})
	require.Same(t, s, GetSuggestionsChanges())
	s.SetCurrentExperiment(&SuggestionChangesExperiment{Name: "A", Parameters: map[string]any{"topN": 5}})
	require.True(t, IsRunningExperiment("A"))
	require.False(t, IsRunningExperiment("B"))

	s.RecordCorrect()
	s.RecordNotFound()
	s.RecordSuggestionPos(2)
	require.Equal(t, 1, s.Correct["A"])
	require.Equal(t, 1, s.NotFound["A"])
	require.Equal(t, 2, s.PosSum["A"])
	require.Equal(t, 3, s.NumSamples["A"])
}
