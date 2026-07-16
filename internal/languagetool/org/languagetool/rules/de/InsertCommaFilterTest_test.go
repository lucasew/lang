package de

// Twin of InsertCommaFilterTest.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertCommaFilter_Filter(t *testing.T) {
	f := NewInsertCommaFilter()
	// two tokens
	require.Equal(t, []string{"hoffe, es"}, f.Suggest("hoffe es"))
	// three tokens: both common placements
	s := f.Suggest("Ich hoffe es")
	require.Contains(t, s, "Ich hoffe, es")
	require.Contains(t, s, "Ich, hoffe es")
	s2 := f.Suggest("Sag mal hast")
	require.Contains(t, s2, "Sag mal, hast")
}
