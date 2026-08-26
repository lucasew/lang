package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExample_WrongAndFixed(t *testing.T) {
	w := Wrong("This is <marker>wrng</marker>.")
	require.Equal(t, "This is <marker>wrng</marker>.", w.GetExample())
	require.Empty(t, w.GetCorrections())
	require.Panics(t, func() { Wrong("no marker") })

	c := Fixed("This is correct.")
	require.Equal(t, "This is correct.", c.GetExample())
	require.Equal(t, "hello", CleanMarkersInExample("<marker>hello</marker>"))
}

func TestIncorrectExample_Corrections(t *testing.T) {
	ex := NewIncorrectExample("a <marker>b</marker>", "c", "d")
	require.Equal(t, []string{"c", "d"}, ex.GetCorrections())
}

func TestExampleSentence_UnbalancedMarkers(t *testing.T) {
	require.Panics(t, func() { NewExampleSentence("<marker>only") })
	require.Panics(t, func() { NewExampleSentence("only</marker>") })
}
