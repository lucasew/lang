package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMoreVariants(t *testing.T) {
	require.Equal(t, "Italian", Italian.Name)
	require.Equal(t, "Dutch", DutchNetherlands.GetName())
	require.Len(t, AllDutchVariants(), 2)
	require.Equal(t, "Polish", Polish.Name)
	require.True(t, ValencianCatalan.Valencian)
	require.True(t, BalearicCatalan.Balearic)
	require.Equal(t, "balear", BalearicCatalan.GetVariant())
	require.Equal(t, "valencia", ValencianCatalan.GetVariant())
	require.Equal(t, "Russian", Russian.Name)
	require.Len(t, AllCatalanVariants(), 3)
	require.True(t, IsNonSwissGerman("de-DE"))
	require.False(t, IsNonSwissGerman("de-CH"))
}

func TestDefaultSpellingRuleIDs_MatchJavaGetId(t *testing.T) {
	// Faithful createDefaultSpellingRule / Morfologik*SpellerRule.getId ports.
	require.Equal(t, "MORFOLOGIK_RULE_IT_IT", Italian.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_NL_NL", DutchNetherlands.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_NL_NL", BelgianDutch.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_PL_PL", Polish.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_CA_ES", Catalan.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_CA_ES", ValencianCatalan.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_RU_RU", Russian.GetDefaultSpellingRuleID())
	require.Equal(t, "FR_SPELLING_RULE", FrenchFrance.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_ES", SpanishSpain.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", PortugalPortuguese.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", BrazilianPortuguese.GetDefaultSpellingRuleID())
}

func TestDutch_GetRuleFileNames(t *testing.T) {
	// Dutch adds nl/nl-NL/grammar.xml; BelgianDutch removes it.
	exists := func(string) bool { return false }
	nl := DutchNetherlands.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/nl/grammar.xml",
		nlNLGrammarXML,
	}, nl)
	be := BelgianDutch.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{"/org/languagetool/rules/nl/grammar.xml"}, be)
	require.NotContains(t, be, nlNLGrammarXML)
}

func TestMaintainers_MatchJava(t *testing.T) {
	require.Equal(t, "Mike Unwalla", AmericanEnglish.GetMaintainers()[0].Name)
	require.Equal(t, MarcinMilkowski.Name, AmericanEnglish.GetMaintainers()[1].Name)
	require.Equal(t, DominiquePelle.Name, FrenchFrance.GetMaintainers()[0].Name)
	require.Equal(t, "Jaume Ortolà", SpanishSpain.GetMaintainers()[0].Name)
	require.Equal(t, "Marco A.G. Pinto", PortugalPortuguese.GetMaintainers()[0].Name)
	require.Equal(t, "OpenTaal", DutchNetherlands.GetMaintainers()[0].Name)
	require.Equal(t, "Paolo Bianchini", Italian.GetMaintainers()[0].Name)
	require.Equal(t, MarcinMilkowski.Name, Polish.GetMaintainers()[0].Name)
	require.Equal(t, "Ricard Roca", Catalan.GetMaintainers()[0].Name)
	require.Equal(t, "Yakov Reztsov", Russian.GetMaintainers()[0].Name)
	require.Equal(t, "Золтан Чала (Csala, Zoltán)", DefaultSerbian.GetMaintainers()[0].Name)
	require.Equal(t, "Alex Buloichik", Belarusian.GetMaintainers()[0].Name)
	require.Equal(t, "Zdenko Podobný", Slovak.GetMaintainers()[0].Name)
}

func TestVariantDefaultRules(t *testing.T) {
	// ValencianCatalan
	require.Contains(t, ValencianCatalan.GetDefaultEnabledRulesForVariant(), "EXIGEIX_VERBS_VALENCIANS")
	require.Contains(t, ValencianCatalan.GetDefaultDisabledRulesForVariant(), "EXIGEIX_VERBS_CENTRAL")
	require.Contains(t, ValencianCatalan.GetDefaultDisabledRulesForVariant(), "CASTIC_CASTIG")
	// BalearicCatalan
	require.Equal(t, []string{"EXIGEIX_VERBS_BALEARS"}, BalearicCatalan.GetDefaultEnabledRulesForVariant())
	require.Equal(t, []string{"EXIGEIX_VERBS_CENTRAL", "CA_SIMPLE_REPLACE_BALEARIC"}, BalearicCatalan.GetDefaultDisabledRulesForVariant())
	// base Catalan empty
	require.Empty(t, Catalan.GetDefaultEnabledRulesForVariant())
	require.Empty(t, Catalan.GetDefaultDisabledRulesForVariant())
	// French BE/CH/CA disable DOUBLER_UNE_CLASSE
	require.Equal(t, []string{"DOUBLER_UNE_CLASSE"}, BelgianFrench.GetDefaultDisabledRulesForVariant())
	require.Equal(t, []string{"DOUBLER_UNE_CLASSE"}, SwissFrench.GetDefaultDisabledRulesForVariant())
	require.Equal(t, []string{"DOUBLER_UNE_CLASSE"}, CanadianFrench.GetDefaultDisabledRulesForVariant())
	require.Empty(t, FrenchFrance.GetDefaultDisabledRulesForVariant())
	// Spanish voseo
	require.Equal(t, []string{"VOSEO"}, SpanishVoseo.GetDefaultDisabledRulesForVariant())
	require.Empty(t, SpanishSpain.GetDefaultDisabledRulesForVariant())
	// code lookup
	en, dis := VariantDefaultRulesForCode("ca-ES-valencia")
	require.Contains(t, en, "EXIGEIX_VERBS_VALENCIANS")
	require.Contains(t, dis, "EXIGEIX_VERBS_CENTRAL")
	_, disBE := VariantDefaultRulesForCode("fr-BE")
	require.Equal(t, []string{"DOUBLER_UNE_CLASSE"}, disBE)
	_, disVoseo := VariantDefaultRulesForCode("es-AR")
	require.Equal(t, []string{"VOSEO"}, disVoseo)
	enEmpty, disEmpty := VariantDefaultRulesForCode("ca")
	require.Empty(t, enEmpty)
	require.Empty(t, disEmpty)
}

func TestNewJLanguageTool_VariantDefaults(t *testing.T) {
	// Blank import path: language init wires VariantDefaultRulesHook.
	// Valencian enables EXIGEIX_VERBS_VALENCIANS; BE French disables DOUBLER_UNE_CLASSE.
	// Import cycle free: this package is language itself — call apply via New is in parent.
	// Test VariantDefaultRulesForCode wiring only here.
	en, dis := VariantDefaultRulesForCode("ca-ES-valencia")
	require.Contains(t, en, "EXIGEIX_VERBS_VALENCIANS")
	require.Contains(t, dis, "EXIGEIX_VERBS_CENTRAL")
}

func TestDutchRelevantRuleIDs(t *testing.T) {
	ids := DutchRelevantRuleIDs()
	require.Contains(t, ids, "MORFOLOGIK_RULE_NL_NL")
	require.Contains(t, ids, "NL_COMPOUNDS")
	require.Contains(t, ids, "NL_SPACE_IN_COMPOUND")
	require.Contains(t, ids, "NL_CHECKCASE")
	require.Equal(t, ids, DutchNetherlands.GetRelevantRuleIDs())
	require.Equal(t, ids, BelgianDutch.GetRelevantRuleIDs())
}

func TestFrenchRelevantRuleIDs(t *testing.T) {
	ids := FrenchRelevantRuleIDs()
	require.Contains(t, ids, "FR_SPELLING_RULE")
	require.Contains(t, ids, "FRENCH_WHITESPACE")
	require.Contains(t, ids, "FRENCH_WHITESPACE_STRICT")
	require.Contains(t, ids, "FR_REPEATEDWORDS")
	require.Equal(t, ids, FrenchFrance.GetRelevantRuleIDs())
}

func TestSpanishRelevantRuleIDs(t *testing.T) {
	ids := SpanishRelevantRuleIDs()
	require.Contains(t, ids, "ES_QUESTION_MARK")
	require.Contains(t, ids, "MORFOLOGIK_RULE_ES")
	require.Contains(t, ids, "ES_SIMPLE_REPLACE_SIMPLE")
	require.Contains(t, ids, "ES_REPEATEDWORDS")
	require.Equal(t, ids, SpanishSpain.GetRelevantRuleIDs())
}

func TestItalianRelevantRuleIDs(t *testing.T) {
	ids := ItalianRelevantRuleIDs()
	require.Contains(t, ids, "WHITESPACE_PUNCTUATION")
	require.Contains(t, ids, "MORFOLOGIK_RULE_IT_IT")
	require.Contains(t, ids, "ITALIAN_WORD_REPEAT_RULE")
	require.Equal(t, ids, Italian.GetRelevantRuleIDs())
}

func TestPortuguesePolishCatalanRussianRelevantRuleIDs(t *testing.T) {
	// Portuguese variant speller IDs
	pt := PortugalPortuguese.GetRelevantRuleIDs()
	require.Contains(t, pt, "MORFOLOGIK_RULE_PT_PT")
	require.Contains(t, pt, "PT_COMPOUNDS_POST_REFORM")
	require.Contains(t, pt, "FILLER_WORDS_PT")
	require.Contains(t, pt, "READABILITY_RULE_SIMPLE_PT")
	require.Contains(t, pt, "READABILITY_RULE_DIFFICULT_PT")
	// Java comments out archaisms (#3095) and weasel-words in getRelevantRules
	require.NotContains(t, pt, "PT_ARCHAISMS_REPLACE")
	require.NotContains(t, pt, "PT_WEASELWORD_REPLACE")
	br := BrazilianPortuguese.GetRelevantRuleIDs()
	require.Contains(t, br, "MORFOLOGIK_RULE_PT_BR")
	require.NotContains(t, br, "MORFOLOGIK_RULE_PT_PT")

	pl := PolishRelevantRuleIDs()
	require.Contains(t, pl, "MORFOLOGIK_RULE_PL_PL")
	require.Contains(t, pl, "DASH_RULE")
	require.Contains(t, pl, "PL_COMPOUNDS")
	require.Equal(t, pl, Polish.GetRelevantRuleIDs())

	ca := CatalanRelevantRuleIDs()
	require.Contains(t, ca, "MORFOLOGIK_RULE_CA_ES")
	require.Contains(t, ca, "IGNORE_PROPER_NOUNS")
	require.Contains(t, ca, "CA_SIMPLE_REPLACE_BALEARIC")
	// Java CatalanUnpairedBracketsRule getId override is commented → UNPAIRED_BRACKETS
	require.Contains(t, ca, "UNPAIRED_BRACKETS")
	require.NotContains(t, ca, "CA_UNPAIRED_BRACKETS")
	require.Contains(t, ca, "CA_UNPAIRED_QUESTION")
	require.Contains(t, ca, "CA_UNPAIRED_EXCLAMATION")
	require.NotContains(t, ca, "CA_REPEATEDWORDS")
	require.Equal(t, ca, Catalan.GetRelevantRuleIDs())
	val := ValencianCatalan.GetRelevantRuleIDs()
	require.Contains(t, val, "CA_WORD_COHERENCY_VALENCIA")
	require.Equal(t, len(ca)+1, len(val))

	ru := RussianRelevantRuleIDs()
	require.Contains(t, ru, "MORFOLOGIK_RULE_RU_RU")
	require.Contains(t, ru, "MORFOLOGIK_RULE_RU_RU_YO")
	require.Contains(t, ru, "FILLER_WORDS_RU")
	require.NotContains(t, ru, "DOUBLE_PUNCTUATION") // commented out in Java
	require.NotContains(t, ru, "READABILITY_RULE_SIMPLE")
	require.Equal(t, ru, Russian.GetRelevantRuleIDs())
}
