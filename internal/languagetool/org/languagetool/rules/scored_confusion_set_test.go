package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScoredConfusionSet(t *testing.T) {
	desc := "there"
	s := NewScoredConfusionSet(1.5, []*ConfusionString{
		NewConfusionString("their", &desc),
		NewConfusionString("there", nil),
	})
	require.Equal(t, float32(1.5), s.GetScore())
	require.Equal(t, []string{"their", "there"}, s.GetConfusionTokens())
	require.Panics(t, func() { NewScoredConfusionSet(0, nil) })
}
