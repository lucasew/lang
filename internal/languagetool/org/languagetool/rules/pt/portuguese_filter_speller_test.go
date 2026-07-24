package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTryWirePortugueseFilterSpeller(t *testing.T) {
	ClearPortugueseFilterSpeller()
	t.Cleanup(ClearPortugueseFilterSpeller)
	// May skip if no spelling dict in tree — still must not panic.
	wired := TryWirePortugueseFilterSpeller()
	if !wired {
		t.Skip("no pt/spelling/*.dict in tree (Java MorfologikPortugueseSpellerRule)")
	}
	require.True(t, FilterDictAvailable())
	// junk should be misspelled when dict is real
	require.True(t, FilterDictIsMisspelled("xyzzyqqqnotaword"))
}

func TestDiscoverAndLoadPortugueseMultitokenSpeller_Official(t *testing.T) {
	sp := DiscoverAndLoadPortugueseMultitokenSpeller()
	require.NotNil(t, sp)
	require.NotNil(t, sp.MultitokenSpeller)
}
