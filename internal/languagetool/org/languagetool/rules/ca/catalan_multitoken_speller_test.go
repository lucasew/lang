package ca

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanMultitokenSpeller(t *testing.T) {
	s := NewCatalanMultitokenSpeller()
	require.NoError(t, s.LoadWords(strings.NewReader("Sant Cugat\n")))
	require.Contains(t, s.GetSuggestions("sant cugat"), "Sant Cugat")
	require.NotEmpty(t, CatalanMultitokenResourcePaths)
}
