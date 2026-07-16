package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDonarseliBeFilter(t *testing.T) {
	f := NewDonarseliBeFilter()
	require.True(t, IsAdverbiFinal("bé"))
	require.Equal(t, "malament", NormalizeAdverbi("mal"))
	require.True(t, IsPronomPersonal("mi"))
	require.True(t, IsExceptionQue("ja"))
	got := f.BuildDonarSuggestion("li", "dona", true, "")
	require.Contains(t, got, "dona")
	require.NotEmpty(t, got)
}
