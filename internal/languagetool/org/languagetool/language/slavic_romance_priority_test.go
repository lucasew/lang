package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGalicianPriorityForId(t *testing.T) {
	require.Equal(t, 19, len(GalicianPriorityExactMap()))
	require.Equal(t, 30, GalicianPriorityForId("DEGREE_MINUTES_SECONDS"))
	require.Equal(t, -5, GalicianPriorityForId("UNPAIRED_BRACKETS"))
	require.Equal(t, -10, GalicianPriorityForId("GL_BARBARISM_REPLACE"))
	require.Equal(t, -50, GalicianPriorityForId("HUNSPELL_RULE"))
	require.Equal(t, -210, GalicianPriorityForId("REPEATED_WORDS"))
	require.Equal(t, -1000, GalicianPriorityForId("TOO_LONG_SENTENCE_35"))
	require.Equal(t, -1004, GalicianPriorityForId("TOO_LONG_SENTENCE_60"))
	// commented Java case FRAGMENT_TWO_ARTICLES must not invent
	require.Equal(t, 0, GalicianPriorityForId("FRAGMENT_TWO_ARTICLES"))
	require.Equal(t, -50, GalicianPriorityForId("SOME_STYLE_RULE"))
}

func TestRussianPriorityForId(t *testing.T) {
	require.Equal(t, 9, len(RussianPriorityExactMap()))
	require.Equal(t, 12, RussianPriorityForId("RU_DASH_RULE"))
	require.Equal(t, 11, RussianPriorityForId("RU_COMPOUNDS"))
	require.Equal(t, 10, RussianPriorityForId("RUSSIAN_SIMPLE_REPLACE_RULE"))
	require.Equal(t, 9, RussianPriorityForId("RUSSIAN_SPECIFIC_CASE"))
	require.Equal(t, 2, RussianPriorityForId("MORFOLOGIC_RULE_RU_RU_YO"))
	require.Equal(t, 1, RussianPriorityForId("MORFOLOGIC_RULE_RU_RU"))
	require.Equal(t, -1, RussianPriorityForId("Word_root_repeat"))
	require.Equal(t, -15, RussianPriorityForId("TOO_LONG_PARAGRAPH"))
	require.Equal(t, 0, RussianPriorityForId("UNKNOWN_RU_XYZ"))
}

func TestBelarusianPriorityForId(t *testing.T) {
	require.Equal(t, 5, len(BelarusianPriorityExactMap()))
	require.Equal(t, 10, BelarusianPriorityForId("RUSSIAN_SIMPLE_REPLACE_RULE"))
	require.Equal(t, 9, BelarusianPriorityForId("BELARUSIAN_SPECIFIC_CASE"))
	require.Equal(t, -1, BelarusianPriorityForId("Word_root_repeat"))
	require.Equal(t, -15, BelarusianPriorityForId("TOO_LONG_PARAGRAPH"))
}

func TestIrishPriorityForId(t *testing.T) {
	require.Equal(t, -15, IrishPriorityForId("TOO_LONG_PARAGRAPH"))
	// base Language still applies for STYLE
	require.Equal(t, -50, IrishPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, IrishPriorityForId("UNKNOWN_GA_XYZ"))
}

func TestPolishPriorityForId(t *testing.T) {
	require.Equal(t, -1, PolishPriorityForId("ZDANIA_ZLOZONE"))
	require.Equal(t, -50, PolishPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, PolishPriorityForId("UNKNOWN_PL_XYZ"))
}
