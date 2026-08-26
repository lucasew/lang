package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugalPortuguesePriorityForId(t *testing.T) {
	// Java PortugalPortuguese.id2prio
	require.Equal(t, 1, PortugalPortuguesePriorityForId("PT_COMPOUNDS_POST_REFORM"))
	require.Equal(t, -9, PortugalPortuguesePriorityForId("PORTUGUESE_OLD_SPELLING_INTERNAL"))
	// super Portuguese prefix gates still apply
	require.Equal(t, -50, PortugalPortuguesePriorityForId("MORFOLOGIK_RULE_PT_PT"))
	require.Equal(t, 5, PortugalPortuguesePriorityForId("HOMOPHONE_AS_CARD"))
	require.Equal(t, 2, len(PortugalPortuguesePriorityMap()))
}

func TestPortuguesePriorityForIdForCode(t *testing.T) {
	ptPT := PortuguesePriorityForIdForCode("pt-PT")
	require.Equal(t, 1, ptPT("PT_COMPOUNDS_POST_REFORM"))
	// Java default variant is pt-PT
	pt := PortuguesePriorityForIdForCode("pt")
	require.Equal(t, -9, pt("PORTUGUESE_OLD_SPELLING_INTERNAL"))
	// Brazilian: base Portuguese map only (Java BrazilianPortuguese has no extra id2prio).
	// Portuguese.id2prio has PT_COMPOUNDS_POST_REFORM: -45; PortugalPortuguese overrides to 1.
	br := PortuguesePriorityForIdForCode("pt-BR")
	require.Equal(t, -45, br("PT_COMPOUNDS_POST_REFORM"))
	require.Equal(t, -50, br("MORFOLOGIK_RULE_PT_BR"))
	require.True(t, isPortugalPortugueseCode("pt-PT"))
	require.False(t, isPortugalPortugueseCode("pt-BR"))
	require.False(t, isPortugalPortugueseCode("pt-AO"))
	require.True(t, isBrazilianPortugueseCode("pt-BR"))
	require.False(t, isBrazilianPortugueseCode("pt-PT"))
	require.False(t, isBrazilianPortugueseCode("pt"))
	require.False(t, isBrazilianPortugueseCode("pt-AO"))
	require.True(t, isAngolaPortugueseCode("pt-AO"))
	require.True(t, isAngolaPortugueseCode("pt_ao"))
	require.False(t, isAngolaPortugueseCode("pt-PT"))
	require.False(t, isAngolaPortugueseCode("pt-BR"))
	require.False(t, isAngolaPortugueseCode("pt"))
	require.False(t, isAngolaPortugueseCode("pt-MZ"))
	require.True(t, isMozambiquePortugueseCode("pt-MZ"))
	require.True(t, isMozambiquePortugueseCode("pt_mz"))
	require.False(t, isMozambiquePortugueseCode("pt-PT"))
	require.False(t, isMozambiquePortugueseCode("pt-BR"))
	require.False(t, isMozambiquePortugueseCode("pt-AO"))
	require.False(t, isMozambiquePortugueseCode("pt"))
}
