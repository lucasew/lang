package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnarASuggestionsFilter(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	f.SynthFuturePresent = func(lemma, pn string) []string {
		return []string{"farem", "fem"}
	}
	got := f.Suggest("fer", "1P0.", "li ho", "anem")
	require.Len(t, got, 2)
	require.Contains(t, got[0], "farem")
}
