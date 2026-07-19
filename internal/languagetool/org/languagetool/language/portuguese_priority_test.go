package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortuguesePriorityMap_Size(t *testing.T) {
	m := PortuguesePriorityMap()
	require.Equal(t, 62, len(m))
	// defensive copy
	if _, ok := m["HOMOPHONE"]; ok {
		m["HOMOPHONE"] = 999
		require.NotEqual(t, 999, PortuguesePriorityForId("HOMOPHONE"))
	}
}

func TestPortuguesePriorityForId_PrefixesBeforeMap(t *testing.T) {
	// Java: these prefixes BEFORE id2prio lookup
	require.Equal(t, -50, PortuguesePriorityForId("MORFOLOGIK_RULE_PT_PT"))
	require.Equal(t, -49, PortuguesePriorityForId("PT_SIMPLE_REPLACE_ORTHOGRAPHY_X"))
	require.Equal(t, -48, PortuguesePriorityForId("AI_PT_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL_X"))
	require.Equal(t, -48, PortuguesePriorityForId("PT_MULTITOKEN_SPELLING_X"))
	require.Equal(t, -4, PortuguesePriorityForId("AI_PT_GGEC_REPLACEMENT_OTHER_X"))
	require.Equal(t, -51, PortuguesePriorityForId("ACENTUAÇÃO_VOGAL_ÊNCLISE_X"))
	require.Equal(t, -52, PortuguesePriorityForId("COLOCACAO_PRONOMINAL_COM_ATRATOR_X"))
	require.Equal(t, -51, PortuguesePriorityForId("AI_PT_HYDRA_LEO_MISSING_COMMA_X"))
	require.Equal(t, -51, PortuguesePriorityForId("AI_PT_HYDRA_LEO_OTHER"))
}

func TestPortuguesePriorityForId_MapSpotChecks(t *testing.T) {
	// Java id2prio keys that are not swallowed by prefix gates — not invent
	require.Equal(t, 5, PortuguesePriorityForId("HOMOPHONE_AS_CARD"))
	require.Equal(t, 10, PortuguesePriorityForId("CONFUSION_POR_PÔR_V2"))
	require.Equal(t, 30, PortuguesePriorityForId("DEGREE_MINUTES_SECONDS"))
	require.Equal(t, -26, PortuguesePriorityForId("ARCHAISMS"))
	require.Equal(t, -90, PortuguesePriorityForId("COLOCAÇÃO_ADVÉRBIO"))
	require.Equal(t, -50, PortuguesePriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, PortuguesePriorityForId("COMPLETELY_UNKNOWN_PT_XYZ"))
}

func TestPortuguesePrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"casa"}, PortuguesePrepareLineForSpeller("casa\tNCMS000"))
	require.Equal(t, []string{"foo"}, PortuguesePrepareLineForSpeller("foo;_Latin_"))
	require.Equal(t, []string{""}, PortuguesePrepareLineForSpeller("ver\tVMIP3S0"))
	require.Equal(t, []string{"plain"}, PortuguesePrepareLineForSpeller("plain"))
	require.Equal(t, []string{"casa"}, PortuguesePrepareLineForSpeller("casa\tN#c"))
}
