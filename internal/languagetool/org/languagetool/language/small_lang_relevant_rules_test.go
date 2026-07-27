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

	// Greek is GreekLang — full twin in greek_test.go

	// Irish is IrishLang — full twin in irish_test.go

	// Esperanto is EsperantoLang — full twin in esperanto_test.go
	// Danish is DanishLang — full twin in danish_test.go
	// Belarusian is BelarusianLang — full twin in belarusian_test.go
	// Breton is BretonLang — full twin in breton_test.go
	// Romanian is RomanianLang — full twin in romanian_test.go
	// Slovak is SlovakLang — full twin in slovak_test.go

	// Japanese / Chinese minimal lists
	require.Equal(t, []string{"DOUBLE_PUNCTUATION", "WHITESPACE_RULE"}, Japanese.GetRelevantRuleIDs())
	require.Equal(t, []string{"DOUBLE_PUNCTUATION", "WHITESPACE_RULE"}, Chinese.GetRelevantRuleIDs())

	// Extended small modules
	require.Contains(t, Khmer.GetRelevantRuleIDs(), "KM_SIMPLE_REPLACE")
	require.Contains(t, Tamil.GetRelevantRuleIDs(), "TOO_LONG_SENTENCE")
	// Tagalog is TagalogLang — full twin in tagalog_test.go
	require.Contains(t, Icelandic.GetRelevantRuleIDs(), "HUNSPELL_NO_SUGGEST_RULE")
	require.Contains(t, Malayalam.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_ML_IN")
	// Persian is PersianLang — full twin in persian_test.go
	require.Contains(t, Lithuanian.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_LT_LT")
	require.Contains(t, CrimeanTatar.GetRelevantRuleIDs(), "MORFOLOGIK_RULE_CRH_UA")
	// Asturian is AsturianLang — full twin in asturian_test.go
	// Slovenian is SlovenianLang — full twin in slovenian_test.go

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
