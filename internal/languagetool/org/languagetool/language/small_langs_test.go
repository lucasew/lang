package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmallLangs(t *testing.T) {
	require.GreaterOrEqual(t, len(AllSmallLangs()), 12)
	require.Equal(t, "Ukrainian", Ukrainian.GetName())
	require.Equal(t, "sk", Slovak.GetShortCode())
}

func TestSmallLang_DefaultSpellingRuleIDs_MatchJava(t *testing.T) {
	// Faithful createDefaultSpellingRule / registered speller getId only.
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", Slovak.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Danish.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Swedish.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Esperanto.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Galician.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Khmer.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_NO_SUGGEST_RULE", Icelandic.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", Irish.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_BR_FR", Breton.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_CRH_UA", CrimeanTatar.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", Slovenian.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_LT_LT", Lithuanian.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_ML_IN", Malayalam.GetDefaultSpellingRuleID())
	// No invent: languages without a Java default speller
	require.Empty(t, Japanese.GetDefaultSpellingRuleID())
	require.Empty(t, Chinese.GetDefaultSpellingRuleID())
	require.Empty(t, Persian.GetDefaultSpellingRuleID())
	require.Empty(t, Tamil.GetDefaultSpellingRuleID())
}

func TestBelarusian_IgnoredCharactersAndSpellerID(t *testing.T) {
	// Java MorfologikBelarusianSpellerRule.getId / createDefaultSpellingRule
	require.Equal(t, "MORFOLOGIK_RULE_BE_BY", Belarusian.GetDefaultSpellingRuleID())
	// Java Belarusian.getIgnoredCharactersRegex: soft hyphen + combining acute/grave
	re := Belarusian.GetIgnoredCharactersRegex()
	require.NotNil(t, re)
	require.True(t, re.MatchString("\u00AD"))
	require.True(t, re.MatchString("\u0301"))
	require.True(t, re.MatchString("\u0300"))
}

func TestSmallLang_GetMaintainers(t *testing.T) {
	// Sample of Java getMaintainers ports
	require.Equal(t, "Leif-Jöran Olsson", Swedish.GetMaintainers()[0].Name)
	require.Equal(t, "Panagiotis Minos", Greek.GetMaintainers()[0].Name)
	require.Equal(t, DominiquePelle.Name, Esperanto.GetMaintainers()[0].Name)
	require.Equal(t, DominiquePelle.Name, Breton.GetMaintainers()[0].Name)
	require.Equal(t, "Fulup Jakez", Breton.GetMaintainers()[1].Name)
	require.Equal(t, "Jim O'Regan", Irish.GetMaintainers()[0].Name)
	require.Len(t, Irish.GetMaintainers(), 4)
	require.Equal(t, "Susana Sotelo Docío", Galician.GetMaintainers()[0].Name)
	require.Equal(t, "Andriy Rysin", CrimeanTatar.GetMaintainers()[0].Name)
	require.Equal(t, "Andriy Rysin", Ukrainian.GetMaintainers()[0].Name)
	require.Equal(t, "Maksym Davydov", Ukrainian.GetMaintainers()[1].Name)
	require.Equal(t, "Elanjelian Venugopal", Tamil.GetMaintainers()[0].Name)
	require.Equal(t, "Nathaniel Oco", Tagalog.GetMaintainers()[0].Name)
	require.Equal(t, "Tao Lin", Chinese.GetMaintainers()[0].Name)
	require.Equal(t, "Taha Zerrouki", Arabic.GetMaintainers()[0].Name)
	require.Equal(t, "Sohaib Afifi", Arabic.GetMaintainers()[1].Name)
	// multi-country Arabic → short code only
	require.Equal(t, "ar", Arabic.GetShortCodeWithCountryAndVariant())
}

func TestUkrainianRelevantRuleIDs_Order(t *testing.T) {
	// Java Arrays.asList order: comma … before speller, speller before hyphen
	ids := UkrainianRelevantRuleIDs()
	require.Equal(t, "COMMA_PARENTHESIS_WHITESPACE", ids[0])
	require.Equal(t, "UPPERCASE_SENTENCE_START", ids[1])
	require.Equal(t, "WHITESPACE_RULE", ids[2])
	require.Equal(t, "UKRAINIAN_WORD_REPEAT_RULE", ids[3])
	require.Equal(t, "DASH", ids[4])
	require.Equal(t, "UK_HIDDEN_CHARS", ids[5])
	require.Equal(t, "MORFOLOGIK_RULE_UK_UA", ids[6])
	require.Equal(t, "UK_MISSING_HYPHEN", ids[7])
	require.Equal(t, "UK_SIMPLE_REPLACE", ids[len(ids)-1])
	require.Equal(t, ids, UkrainianLanguageDefault.GetRelevantRuleIDs())
	ms := UkrainianLanguageDefault.GetMaintainers()
	require.Equal(t, "Andriy Rysin", ms[0].Name)
	require.Equal(t, "Maksym Davydov", ms[1].Name)
}

func TestSlovak_GetRuleFileNames(t *testing.T) {
	// Java Slovak.RULE_FILES → grammar-typography.xml after base grammar
	exists := func(string) bool { return false }
	files := Slovak.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/sk/grammar.xml",
		"/org/languagetool/rules/sk/grammar-typography.xml",
	}, files)
}

func TestUkrainianSmallLang_GetRuleFileNames(t *testing.T) {
	exists := func(string) bool { return false }
	files := Ukrainian.GetRuleFileNamesWithExists(exists)
	require.Equal(t, []string{
		"/org/languagetool/rules/uk/grammar.xml",
		"/org/languagetool/rules/uk/grammar-spelling.xml",
		"/org/languagetool/rules/uk/grammar-grammar.xml",
		"/org/languagetool/rules/uk/grammar-barbarism.xml",
		"/org/languagetool/rules/uk/grammar-style.xml",
		"/org/languagetool/rules/uk/grammar-punctuation.xml",
	}, files)
}
