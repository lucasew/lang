package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDutchPriorityMap_Size(t *testing.T) {
	m := DutchPriorityMap()
	require.Equal(t, 17, len(m))
	m["ET_AL"] = 999
	require.Equal(t, 1, DutchPriorityForId("ET_AL"))
}

func TestDutchPriorityForId_MapAndPrefixes(t *testing.T) {
	// Java: NL_SIMPLE_REPLACE / NL_SPACE_IN_COMPOUND before map → 1
	require.Equal(t, 1, DutchPriorityForId("NL_SIMPLE_REPLACE"))
	require.Equal(t, 1, DutchPriorityForId("NL_SIMPLE_REPLACE_FOO"))
	require.Equal(t, 1, DutchPriorityForId("NL_SPACE_IN_COMPOUND_X"))
	// map
	require.Equal(t, 3, DutchPriorityForId("SINT_X"))
	require.Equal(t, 1, DutchPriorityForId("ET_AL"))
	require.Equal(t, 1, DutchPriorityForId("N_PERSOONS"))
	// AI_NL_HYDRA_LEO
	require.Equal(t, -51, DutchPriorityForId("AI_NL_HYDRA_LEO_MISSING_COMMA_X"))
	require.Equal(t, -5, DutchPriorityForId("AI_NL_HYDRA_LEO_OTHER"))
	require.Equal(t, -50, DutchPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, DutchPriorityForId("COMPLETELY_UNKNOWN_NL_XYZ"))
}
