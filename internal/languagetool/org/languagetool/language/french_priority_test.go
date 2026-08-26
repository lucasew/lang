package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrenchPriorityMap_Size(t *testing.T) {
	m := FrenchPriorityMap()
	require.Equal(t, 99, len(m))
	m["SA_CA_SE"] = 999
	require.Equal(t, 100, FrenchPriorityForId("SA_CA_SE"))
}

func TestFrenchPriorityForId_MapSpotChecks(t *testing.T) {
	// Java id2prio high / low — not invent
	require.Equal(t, 100, FrenchPriorityForId("AGREEMENT_EXCEPTIONS"))
	require.Equal(t, 100, FrenchPriorityForId("SIL_VOUS_PLAIT"))
	require.Equal(t, 100, FrenchPriorityForId("SA_CA_SE"))
	require.Equal(t, -400, FrenchPriorityForId("TOUT_MAJUSCULES"))
	require.Equal(t, -400, FrenchPriorityForId("FRENCH_WHITESPACE"))
	require.Equal(t, -350, FrenchPriorityForId("FRENCH_WORD_REPEAT_BEGINNING_RULE"))
}

func TestFrenchPriorityForId_Prefixes(t *testing.T) {
	// Java getPriorityForId after map miss
	require.Equal(t, 500, FrenchPriorityForId("FR_COMPOUNDS_X"))
	require.Equal(t, 20, FrenchPriorityForId("CAT_TYPOGRAPHIE"))
	require.Equal(t, 20, FrenchPriorityForId("CAT_TOURS_CRITIQUES"))
	require.Equal(t, 20, FrenchPriorityForId("CAT_HOMONYMES_PARONYMES"))
	require.Equal(t, -5, FrenchPriorityForId("SON"))
	require.Equal(t, -50, FrenchPriorityForId("CAR_SOMETHING"))
	require.Equal(t, -50, FrenchPriorityForId("CONFUSION_RULE_PREMIUM"))
	require.Equal(t, -90, FrenchPriorityForId("FR_MULTITOKEN_SPELLING_X"))
	require.Equal(t, 150, FrenchPriorityForId("FR_SIMPLE_REPLACE_FOO"))
	require.Equal(t, -150, FrenchPriorityForId("grammalecte_foo"))
	require.Equal(t, -101, FrenchPriorityForId("AI_FR_HYDRA_LEO_X"))
	require.Equal(t, -101, FrenchPriorityForId("AI_FR_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL"))
	// other AI_FR_GGEC not special → base 0
	require.Equal(t, 0, FrenchPriorityForId("AI_FR_GGEC_SOMETHING_ELSE"))
	require.Equal(t, -50, FrenchPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, FrenchPriorityForId("COMPLETELY_UNKNOWN_FR_XYZ"))
}

func TestFrenchPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"Paris"}, FrenchPrepareLineForSpeller("Paris\tZ"))
	require.Equal(t, []string{"maison"}, FrenchPrepareLineForSpeller("maison;N"))
	require.Equal(t, []string{"bon"}, FrenchPrepareLineForSpeller("bon\tA"))
	require.Equal(t, []string{""}, FrenchPrepareLineForSpeller("manger\tV"))
	require.Equal(t, []string{""}, FrenchPrepareLineForSpeller("Ho Chi Minh"))
	require.Equal(t, []string{""}, FrenchPrepareLineForSpeller("Ho Chi Minh\tZ"))
	require.Equal(t, []string{"plain"}, FrenchPrepareLineForSpeller("plain"))
	require.Equal(t, []string{"Paris"}, FrenchPrepareLineForSpeller("Paris\tZ#comment"))
	require.True(t, FrenchHasMinMatchesRules())
}
