package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpanishPriorityMap_Size(t *testing.T) {
	m := SpanishPriorityMap()
	require.Equal(t, 55, len(m)) // Java active id2prio puts (commented puts excluded)
	// Defensive copy
	m["TYPOGRAPHY"] = 999
	require.Equal(t, 20, SpanishPriorityForId("TYPOGRAPHY"))
}

func TestSpanishPriorityForId_MapSpotChecks(t *testing.T) {
	// Java id2prio — not invent
	require.Equal(t, 50, SpanishPriorityForId("ES_SIMPLE_REPLACE_MULTIWORDS"))
	require.Equal(t, 50, SpanishPriorityForId("LOS_MAPUCHE"))
	require.Equal(t, 20, SpanishPriorityForId("TYPOGRAPHY"))
	require.Equal(t, 15, SpanishPriorityForId("AGREEMENT_DET_NOUN"))
	require.Equal(t, -100, SpanishPriorityForId("MORFOLOGIK_RULE_ES"))
	require.Equal(t, -150, SpanishPriorityForId("SPANISH_WORD_REPEAT_RULE"))
	require.Equal(t, -200, SpanishPriorityForId("UPPERCASE_SENTENCE_START"))
	require.Equal(t, -250, SpanishPriorityForId("ES_QUESTION_MARK"))
	require.Equal(t, -50, SpanishPriorityForId("REPETITIONS_STYLE"))
	require.Equal(t, -40, SpanishPriorityForId("VOSEO"))
}

func TestSpanishPriorityForId_Prefixes(t *testing.T) {
	// Java getPriorityForId special cases before / after map
	require.Equal(t, 50, SpanishPriorityForId("CONFUSIONS2"))
	require.Equal(t, 50, SpanishPriorityForId("RARE_WORDS"))
	require.Equal(t, 40, SpanishPriorityForId("MISSPELLING"))
	require.Equal(t, 40, SpanishPriorityForId("CONFUSIONS"))
	require.Equal(t, 40, SpanishPriorityForId("INCORRECT_EXPRESSIONS"))
	require.Equal(t, 30, SpanishPriorityForId("DIACRITICS"))
	require.Equal(t, 30, SpanishPriorityForId("ES_SIMPLE_REPLACE_SIMPLE_FOO"))
	require.Equal(t, 50, SpanishPriorityForId("ES_COMPOUNDS_BAR"))
	require.Equal(t, -101, SpanishPriorityForId("AI_ES_HYDRA_LEO_X"))
	require.Equal(t, -300, SpanishPriorityForId("AI_ES_GGEC_REPLACEMENT_OTHER"))
	require.Equal(t, 0, SpanishPriorityForId("AI_ES_GGEC_SOMETHING"))
	require.Equal(t, -95, SpanishPriorityForId("ES_MULTITOKEN_SPELLING_X"))
	// base Language fallback
	require.Equal(t, -50, SpanishPriorityForId("SOME_STYLE_RULE"))
	require.Equal(t, 0, SpanishPriorityForId("COMPLETELY_UNKNOWN_ES_XYZ"))
}

func TestSpanishPrepareLineForSpeller(t *testing.T) {
	require.Equal(t, []string{"casa"}, SpanishPrepareLineForSpeller("casa\tNCMS000"))
	require.Equal(t, []string{"casa"}, SpanishPrepareLineForSpeller("casa;NCFS000"))
	require.Equal(t, []string{"foo"}, SpanishPrepareLineForSpeller("foo\t_Latin_"))
	require.Equal(t, []string{"bar"}, SpanishPrepareLineForSpeller("bar\tLOC_ADV"))
	require.Equal(t, []string{""}, SpanishPrepareLineForSpeller("ver\tVMIP3S0"))
	require.Equal(t, []string{"plain"}, SpanishPrepareLineForSpeller("plain"))
	require.Equal(t, []string{"casa"}, SpanishPrepareLineForSpeller("casa\tNCMS000#comment"))
}

func TestSpanishAdaptSuggestion(t *testing.T) {
	// Java ES_CONTRACTIONS: \b([Aa]|[Dd]e) e(l)\b → $1$2
	require.Equal(t, "al libro", SpanishAdaptSuggestion("a el libro", ""))
	require.Equal(t, "del mar", SpanishAdaptSuggestion("de el mar", ""))
	require.Equal(t, "Del", SpanishAdaptSuggestion("De el", ""))
	require.Equal(t, "Al", SpanishAdaptSuggestion("A el", ""))
	require.Equal(t, "casa", SpanishAdaptSuggestion("casa", ""))
	require.True(t, SpanishHasMinMatchesRules())
}
