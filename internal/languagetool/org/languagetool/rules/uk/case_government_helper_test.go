package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseGovernmentHelper(t *testing.T) {
	h := LoadCaseGovernmentHelper()
	require.NotEmpty(t, h.Map)
	require.True(t, h.HasCaseGovernment("згідно з", "v_oru"))
	// sample line from file — find any non-empty non-override entry
	found := false
	for lemma, set := range h.Map {
		if lemma == "згідно з" || len(set) == 0 {
			continue
		}
		found = true
		break
	}
	require.True(t, found, "expected at least one non-empty government map entry")
}
