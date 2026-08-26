package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestedReplacement(t *testing.T) {
	s := NewSuggestedReplacement("colour")
	require.Equal(t, "colour", s.GetReplacement())
	require.Equal(t, SuggestionTypeDefault, s.GetType())
	desc := "British spelling"
	s.SetShortDescription(&desc)
	require.Equal(t, "colour(British spelling)", s.String())

	list := ConvertSuggestions([]string{"a", "b"})
	require.Len(t, list, 2)
	require.Equal(t, "a", list[0].GetReplacement())

	top := TopMatch("foo", nil)
	require.Len(t, top, 1)
	require.NotNil(t, top[0].GetConfidence())
	require.InDelta(t, 0.99, float64(*top[0].GetConfidence()), 0.001)

	cp := CopySuggestedReplacement(s)
	require.Equal(t, s.GetReplacement(), cp.GetReplacement())
	require.Equal(t, *s.GetShortDescription(), *cp.GetShortDescription())
}
