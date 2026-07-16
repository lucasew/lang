package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSynthesizeWithAnyDeterminerFilter(t *testing.T) {
	f := NewSynthesizeWithAnyDeterminerFilter()
	got := f.SuggestAll([]struct{ Form, POS string }{
		{"amic", "NCMS000"},
		{"amiga", "NCFS000"},
	}, "", "MS", "")
	require.Contains(t, got, "l'amic")
	require.Contains(t, got, "l'amiga")
	require.True(t, IsPreposition("de"))
	require.Equal(t, "d", PrepositionKey("de"))
}
