package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseGovernmentHelper(t *testing.T) {
	h := LoadCaseGovernmentHelper()
	require.NotEmpty(t, h.Map)
	require.True(t, h.HasCaseGovernment("згідно з", "v_oru"))
	// sample line from file — first non-comment lemma
	for lemma, set := range h.Map {
		if lemma == "згідно з" {
			continue
		}
		require.NotEmpty(t, set)
		break
	}
}
