package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanAndRomanceVariants(t *testing.T) {
	require.Equal(t, "German (Germany)", GermanyGerman.GetName())
	v, ok := GermanVariantByCode("de-at")
	require.True(t, ok)
	require.Equal(t, "de-AT", v.ShortCode)
	require.Len(t, AllFrenchVariants(), 4)
	require.Len(t, AllSpanishVariants(), 2)
	require.Len(t, AllPortugueseVariants(), 4)
	require.True(t, SpanishVoseo.Voseo)
}

func TestGermanVariant_MaintainedAndShortCode(t *testing.T) {
	require.Equal(t, "de", GermanyGerman.GetShortCode())
	require.Equal(t, "de", SwissGerman.GetShortCode())
	require.Equal(t, languagetool.ActivelyMaintained, GermanyGerman.GetMaintainedState())
	require.True(t, AustrianGerman.IsVariant())
	require.True(t, SwissGerman.IsVariant())
	ms := GermanyGerman.GetMaintainers()
	require.Len(t, ms, 2)
	require.Equal(t, "Jan Schreiber", ms[0].Name)
	require.Equal(t, DanielNaber.Name, ms[1].Name)
	// Quotes: DE „ “ vs CH « »
	require.Equal(t, "„", GermanyGerman.GetOpeningDoubleQuote())
	require.Equal(t, "“", GermanyGerman.GetClosingDoubleQuote())
	require.Equal(t, "«", SwissGerman.GetOpeningDoubleQuote())
	require.Equal(t, "»", SwissGerman.GetClosingDoubleQuote())
	require.Equal(t, "‚", GermanyGerman.GetOpeningSingleQuote())
	require.Equal(t, "‘", GermanyGerman.GetClosingSingleQuote())
}

func TestGermanRelevantRuleIDs(t *testing.T) {
	ids := GermanRelevantRuleIDs()
	require.Contains(t, ids, "DE_AGREEMENT")
	require.Contains(t, ids, "DE_CASE")
	require.Contains(t, ids, "COMMA_IN_FRONT_RELATIVE_CLAUSE")
	require.Contains(t, ids, "COMMA_BEHIND_RELATIVE_CLAUSE")
	require.Contains(t, ids, "READABILITY_RULE_SIMPLE_DE")
	require.Contains(t, ids, "READABILITY_RULE_DIFFICULT_DE")
	require.Contains(t, ids, "DE_REPEATEDWORDS")
	// GermanyGerman + AustrianGerman add GermanCompoundRule (DE_COMPOUNDS)
	de := GermanyGerman.GetRelevantRuleIDs()
	require.Contains(t, de, "DE_COMPOUNDS")
	require.Equal(t, len(ids)+1, len(de))
	at := AustrianGerman.GetRelevantRuleIDs()
	require.Contains(t, at, "DE_COMPOUNDS")
	require.Equal(t, de, at)
	// SwissGerman does not add compound rule
	ch := SwissGerman.GetRelevantRuleIDs()
	require.Equal(t, ids, ch)
	require.NotContains(t, ch, "DE_COMPOUNDS")
}

func TestGermanVariant_GetRuleFileNames(t *testing.T) {
	// Always include base grammar; de-DE-AT for DE/AT; not for CH.
	exists := func(p string) bool { return false } // no style/variant extras
	de := GermanyGerman.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/de/grammar.xml",
		deDEATGrammarXML,
	}, de)
	at := AustrianGerman.GetRuleFileNamesWithExists(exists)
	require.Equal(t, de, at)
	ch := SwissGerman.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{"/org/languagetool/rules/de/grammar.xml"}, ch)
	// Default exists=true may include de/de-DE/grammar.xml etc.
	full := GermanyGerman.GetRuleFileNames()
	require.Contains(t, full, "/org/languagetool/rules/de/grammar.xml")
	require.Contains(t, full, deDEATGrammarXML)
}
