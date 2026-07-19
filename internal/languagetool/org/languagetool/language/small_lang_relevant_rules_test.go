package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmallLangRelevantRuleIDs(t *testing.T) {
	// Ukrainian
	uk := UkrainianRelevantRuleIDs()
	require.Contains(t, uk, "MORFOLOGIK_RULE_UK_UA")
	require.Contains(t, uk, "UK_MIXED_ALPHABETS")
	require.Contains(t, uk, "DASH")
	require.Equal(t, uk, Ukrainian.GetRelevantRuleIDs())

	// Galician
	gl := GalicianRelevantRuleIDs()
	require.Contains(t, gl, "GL_CAST_WORDS")
	require.Contains(t, gl, "HUNSPELL_RULE")
	require.Equal(t, gl, Galician.GetRelevantRuleIDs())

	// Swedish
	require.Contains(t, SwedishRelevantRuleIDs(), "SV_COMPOUNDS")
	require.Equal(t, SwedishRelevantRuleIDs(), Swedish.GetRelevantRuleIDs())

	// Greek — custom unpaired id EL_UNPAIRED_BRACKETS (not UNPAIRED_BRACKETS)
	el := GreekRelevantRuleIDs()
	require.Contains(t, el, "MORFOLOGIK_RULE_EL_GR")
	require.Contains(t, el, "GREEK_HOMONYMS_REPLACE")
	require.Contains(t, el, "EL_UNPAIRED_BRACKETS")
	require.NotContains(t, el, "UNPAIRED_BRACKETS")
	require.Equal(t, el, Greek.GetRelevantRuleIDs())

	// Irish: Java Arrays.asList registers UppercaseSentenceStartRule twice
	ga := IrishRelevantRuleIDs()
	require.Contains(t, ga, "MORFOLOGIK_RULE_GA_IE")
	require.Contains(t, ga, "GA_DHA_NO_BEIRT")
	nUpper := 0
	for _, id := range ga {
		if id == "UPPERCASE_SENTENCE_START" {
			nUpper++
		}
	}
	require.Equal(t, 2, nUpper)
	require.Equal(t, 25, len(ga))
	require.Equal(t, ga, Irish.GetRelevantRuleIDs())

	// Belarusian / Breton / Esperanto / Slovak / Danish / Romanian
	require.Contains(t, BelarusianRelevantRuleIDs(), "MORFOLOGIK_RULE_BE_BY")
	require.Contains(t, BretonRelevantRuleIDs(), "BR_TOPO")
	require.Contains(t, EsperantoRelevantRuleIDs(), "HUNSPELL_RULE")
	require.Contains(t, SlovakRelevantRuleIDs(), "MORFOLOGIK_RULE_SK_SK")
	require.Len(t, DanishRelevantRuleIDs(), 6)
	require.Contains(t, RomanianRelevantRuleIDs(), "MORFOLOGIK_RULE_RO_RO")
	require.Contains(t, RomanianRelevantRuleIDs(), "RO_COMPOUND")

	// Japanese / Chinese minimal lists
	require.Equal(t, []string{"DOUBLE_PUNCTUATION", "WHITESPACE_RULE"}, Japanese.GetRelevantRuleIDs())
	require.Equal(t, []string{"DOUBLE_PUNCTUATION", "WHITESPACE_RULE"}, Chinese.GetRelevantRuleIDs())

	// Extended small modules
	require.Contains(t, Khmer.GetRelevantRuleIDs(), "KM_SIMPLE_REPLACE")
	require.Contains(t, Tamil.GetRelevantRuleIDs(), "TOO_LONG_SENTENCE")
	require.Contains(t, Tagalog.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_TL")
	require.Contains(t, Icelandic.GetRelevantRuleIDs(), "HUNSPELL_NO_SUGGEST_RULE")
	require.Contains(t, Malayalam.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_ML_IN")
	require.Contains(t, Persian.GetRelevantRuleIDs(), "FA_WORD_COHERENCY")
	require.Contains(t, Lithuanian.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_LT_LT")
	require.Contains(t, CrimeanTatar.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_CRH_UA")
	require.Contains(t, Asturian.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_AST")
	require.Contains(t, Slovenian.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_SL_SI")

	// Simple German is not SmallLang — package helper only
	require.Equal(t, []string{"TOO_LONG_SENTENCE_DE"}, SimpleGermanRelevantRuleIDs())
}

func TestSerbianRelevantRuleIDs(t *testing.T) {
	ek := SerbianEkavianRelevantRuleIDs()
	require.Contains(t, ek, "MORFOLOGIK_RULE_SR_EKAVIAN")
	require.Contains(t, ek, "SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE")
	require.Equal(t, ek, DefaultSerbian.GetRelevantRuleIDs())
	jk := SerbianJekavianRelevantRuleIDs()
	require.Contains(t, jk, "MORFOLOGIK_RULE_SR_JEKAVIAN")
	require.Contains(t, jk, "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE")
	// Shared basic prefix
	require.Equal(t, SerbianBasicRelevantRuleIDs(), ek[:len(SerbianBasicRelevantRuleIDs())])
}
