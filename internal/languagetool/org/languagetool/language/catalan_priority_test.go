package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanPriorityExactMap_Size(t *testing.T) {
	m := CatalanPriorityExactMap()
	require.Equal(t, 135, len(m))
	m["CONFUSIONS2"] = 999
	require.Equal(t, 80, CatalanPriorityForId("CONFUSIONS2"))
}

func TestCatalanPriorityForId_ExactSpotChecks(t *testing.T) {
	// Java switch cases — not invent
	require.Equal(t, 80, CatalanPriorityForId("CONFUSIONS2"))
	require.Equal(t, 80, CatalanPriorityForId("DEU_NI_DO"))
	require.Equal(t, 70, CatalanPriorityForId("FER_LOGIN"))
	require.Equal(t, 50, CatalanPriorityForId("INCORRECT_EXPRESSIONS"))
	require.Equal(t, 35, CatalanPriorityForId("ELA_GEMINADA"))
	require.Equal(t, 30, CatalanPriorityForId("CONFUSIONS"))
	require.Equal(t, 20, CatalanPriorityForId("DIACRITICS"))
	require.Equal(t, -100, CatalanPriorityForId("MORFOLOGIK_RULE_CA_ES"))
	require.Equal(t, -120, CatalanPriorityForId("EXIGEIX_ACCENTUACIO_VALENCIANA"))
	require.Equal(t, -150, CatalanPriorityForId("PHRASE_REPETITION"))
	require.Equal(t, -200, CatalanPriorityForId("FALTA_ELEMENT_ENTRE_VERBS"))
	require.Equal(t, -300, CatalanPriorityForId("UPPERCASE_SENTENCE_START"))
	require.Equal(t, -90, CatalanPriorityForId("CA_SPLIT_LONG_SENTENCE"))
	require.Equal(t, -50, CatalanPriorityForId("REPETITIONS_STYLE"))
}

func TestCatalanPriorityForId_Prefixes(t *testing.T) {
	// Java order after switch: more specific prefixes before CA_SIMPLE_REPLACE
	require.Equal(t, -95, CatalanPriorityForId("CA_MULTITOKEN_SPELLING_X"))
	require.Equal(t, 70, CatalanPriorityForId("CA_SIMPLE_REPLACE_MULTIWORDS_X"))
	require.Equal(t, 65, CatalanPriorityForId("CA_SIMPLE_REPLACE_ANGLICISM_X"))
	require.Equal(t, 60, CatalanPriorityForId("CA_SIMPLE_REPLACE_BALEARIC_X"))
	require.Equal(t, 28, CatalanPriorityForId("CA_SIMPLE_REPLACE_VERBS_X"))
	require.Equal(t, 50, CatalanPriorityForId("CA_COMPOUNDS_X"))
	require.Equal(t, 0, CatalanPriorityForId("CA_SIMPLE_REPLACE_DIACRITICS_IEC_X"))
	require.Equal(t, 30, CatalanPriorityForId("CA_SIMPLE_REPLACE_OTHER"))
	// base
	require.Equal(t, -50, CatalanPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, CatalanPriorityForId("COMPLETELY_UNKNOWN_CA_XYZ"))
}

func TestCatalanPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"casa"}, CatalanPrepareLineForSpeller("casa\tNCMS000"))
	require.Equal(t, []string{"foo"}, CatalanPrepareLineForSpeller("foo;_Latin_"))
	require.Equal(t, []string{""}, CatalanPrepareLineForSpeller("ver\tVMIP3S0"))
	require.Equal(t, []string{""}, CatalanPrepareLineForSpeller("Banco Santander"))
	require.Equal(t, []string{""}, CatalanPrepareLineForSpeller("Rosalía\tN"))
	require.Equal(t, []string{"plain"}, CatalanPrepareLineForSpeller("plain"))
}
