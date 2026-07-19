package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanPriorityForId_Map(t *testing.T) {
	// Spot-check Java id2prio entries (not invent).
	require.Equal(t, 10, GermanPriorityForId("OLD_SPELLING_RULE"))
	require.Equal(t, 10, GermanPriorityForId("DE_COMPOUNDS"))
	require.Equal(t, 11, GermanPriorityForId("DE_PROHIBITED_PHRASE"))
	require.Equal(t, -15, GermanPriorityForId("STYLE"))
	require.Equal(t, -15, GermanPriorityForId("REDUNDANCY"))
	// Further Java id2prio keys (German.java static block)
	require.Equal(t, 1, GermanPriorityForId("ALLES_GUTE"))
	require.Equal(t, 10, GermanPriorityForId("WRONG_SPELLING_PREMIUM_INTERNAL"))
	require.Equal(t, 1, GermanPriorityForId("SEIT_VS_SEID"))
	require.Equal(t, -14, GermanPriorityForId("TYPOGRAPHY"))
	require.Equal(t, -15, GermanPriorityForId("COLLOQUIALISMS"))
	require.Equal(t, -15, GermanPriorityForId("GENDER_NEUTRALITY"))
	// AI_DE_MERGED_MATCH is not in id2prio → base 0 (not invent a priority).
	require.Equal(t, 0, GermanPriorityForId("AI_DE_MERGED_MATCH"))
}

func TestGermanPriorityMap_GetPriorityMap(t *testing.T) {
	// Java German.getPriorityMap size + copy semantics.
	m := GermanPriorityMap()
	require.Equal(t, 239, len(m))
	require.Equal(t, 10, m["OLD_SPELLING_RULE"])
	// Defensive copy: mutating return must not change live map.
	m["OLD_SPELLING_RULE"] = 999
	require.Equal(t, 10, GermanPriorityForId("OLD_SPELLING_RULE"))
}

func TestGermanPriorityForId_Prefixes(t *testing.T) {
	require.Equal(t, -4, GermanPriorityForId("DE_PROHIBITED_COMPOUNDS_FOO"))
	require.Equal(t, -2, GermanPriorityForId("DE_MULTITOKEN_SPELLING_X"))
	require.Equal(t, -1, GermanPriorityForId("CONFUSION_RULE_BAR"))
	require.Equal(t, -51, GermanPriorityForId("AI_DE_HYDRA_LEO_MISSING_COMMA_X"))
	require.Equal(t, 2, GermanPriorityForId("AI_DE_HYDRA_LEO_CP_X"))
	require.Equal(t, 1, GermanPriorityForId("AI_DE_HYDRA_LEO_DATAKK_X"))
	require.Equal(t, -11, GermanPriorityForId("AI_DE_HYDRA_LEO_OTHER"))
	require.Equal(t, -52, GermanPriorityForId("AI_DE_KOMMA_X"))
	require.Equal(t, 0, GermanPriorityForId("AI_DE_GGEC_MISSING_PUNCTUATION_E_DASH_MAIL"))
	require.Equal(t, -1, GermanPriorityForId("AI_DE_GGEC_REPLACEMENT_NOUN"))
	require.Equal(t, -2, GermanPriorityForId("AI_DE_GGEC_MISSING_ORTHOGRAPHY_SPACE_X"))
	// Java getPriorityForId AI_DE_GGEC remaining branches
	require.Equal(t, -4, GermanPriorityForId("AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD"))
	require.Equal(t, -4, GermanPriorityForId("AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD_X"))
	require.Equal(t, -2, GermanPriorityForId("AI_DE_GGEC_UNNECESSARY_PUNCTUATION_X"))
	// AI_DE_GGEC_MISSING_PUNCT regex (Java): ..._DASH_J(_|AE)HRIG or REPLACEMENT_CONFUSION → -1
	require.Equal(t, -1, GermanPriorityForId("AI_DE_GGEC_MISSING_PUNCTUATION_1_DASH_JAEHRIG"))
	require.Equal(t, -1, GermanPriorityForId("AI_DE_GGEC_MISSING_PUNCTUATION_2_DASH_J_HRIG"))
	require.Equal(t, -1, GermanPriorityForId("AI_DE_GGEC_REPLACEMENT_CONFUSION"))
	// default AI_DE_GGEC residual
	require.Equal(t, 1, GermanPriorityForId("AI_DE_GGEC_SOMETHING_ELSE"))
}

func TestGermanPriorityForId_BaseFallback(t *testing.T) {
	// Language base: STYLE substring → -50 when not in map (map has STYLE category -15).
	// Unknown style-ish id without map entry:
	require.Equal(t, -50, GermanPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, -101, GermanPriorityForId("TOO_LONG_SENTENCE"))
	// Java id2prio has REPETITIONS_STYLE: -60 (category priority override of base -55).
	require.Equal(t, -60, GermanPriorityForId("REPETITIONS_STYLE"))
	require.Equal(t, 0, GermanPriorityForId("COMPLETELY_UNKNOWN_RULE_XYZ"))
}

func TestGermanAdvancedTypography(t *testing.T) {
	// Port of JLanguageToolTest.testAdvancedTypography (DE).
	require.Equal(t, "Das ist…", GermanAdvancedTypography("Das ist..."))
	require.Equal(t, "Meinten Sie „entschieden“ oder „entscheidend“?",
		GermanAdvancedTypography(`Meinten Sie "entschieden" oder "entscheidend"?`))
	require.Equal(t, "Meinten Sie ‚entschieden‘ oder ‚entscheidend‘?",
		GermanAdvancedTypography("Meinten Sie 'entschieden' oder 'entscheidend'?"))
	require.Equal(t, "z.\u00a0B.", GermanAdvancedTypography("z. B."))
	require.Equal(t, "z.\u00a0B.", GermanAdvancedTypography("z.B."))
	require.Equal(t, "i.\u00a0d.\u00a0R.", GermanAdvancedTypography("i.d.R."))
	require.Equal(t, "i.\u00a0d.\u00a0R.", GermanAdvancedTypography("i. d. R."))
	// Java: nested single around double quote mid-sentence
	require.Equal(t, `Zeichen ohne sein Gegenstück: ‚"‘ scheint zu fehlen`,
		GermanAdvancedTypography(`Zeichen ohne sein Gegenstück: '"' scheint zu fehlen`))
	require.True(t, GermanyGerman.IsAdvancedTypographyEnabled())
	// Java AustrianGerman inherits German quotes (not Swiss « »).
	require.Equal(t, "Meinten Sie „entschieden“?",
		AustrianGerman.ToAdvancedTypography(`Meinten Sie "entschieden"?`))
	require.Equal(t, "z.\u00a0B.", AustrianGerman.ToAdvancedTypography("z.B."))
	require.True(t, AustrianGerman.IsAdvancedTypographyEnabled())
	// GermanyGerman variant API same as package-level GermanAdvancedTypography
	require.Equal(t, GermanAdvancedTypography(`x "y" z`),
		GermanyGerman.ToAdvancedTypography(`x "y" z`))
}

func TestGermanPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"Haus", "Hause"}, GermanPrepareLineForSpeller("Haus/E"))
	require.Equal(t, []string{"Foo", "Fooe", "Foos", "Foon"}, GermanPrepareLineForSpeller("Foo/ESN"))
	// Java: ignore # comments; tags E/S/N expand independently
	require.Equal(t, []string{"Haus", "Hause"}, GermanPrepareLineForSpeller("Haus/E#comment"))
	require.Equal(t, []string{"Bar"}, GermanPrepareLineForSpeller("Bar"))
	require.Equal(t, []string{"Baz", "Bazs"}, GermanPrepareLineForSpeller("Baz/S"))
	require.Equal(t, []string{"Qux", "Quxn"}, GermanPrepareLineForSpeller("Qux/N#x"))
	require.True(t, GermanyGerman.HasMinMatchesRules())
}
