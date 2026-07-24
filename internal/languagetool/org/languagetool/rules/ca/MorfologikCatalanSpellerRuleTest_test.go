package ca

// Twin of MorfologikCatalanSpellerRuleTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikCatalanSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	require.Equal(t, MorfologikCatalanSpellerRuleID, r.GetID())
	require.Equal(t, CatalanSpellerDict, r.GetFileName())
	// Without a loaded dictionary, Match is still safe.
	matches, err := r.Match(nil)
	require.NoError(t, err)
	require.Empty(t, matches)
}
