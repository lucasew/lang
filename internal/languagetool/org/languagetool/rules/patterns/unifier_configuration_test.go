package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnifierConfiguration(t *testing.T) {
	c := NewUnifierConfiguration()
	pt := Token("sg")
	c.SetEquivalence("number", "sg", pt)
	c.SetEquivalence("number", "sg", Token("ignored")) // no-op
	types := c.GetEquivalenceTypes()
	require.Len(t, types, 1)
	require.Equal(t, "sg", types[NewEquivalenceTypeLocator("number", "sg")].Token)
	feats := c.GetEquivalenceFeatures()
	require.Equal(t, []string{"sg"}, feats["number"])
}
