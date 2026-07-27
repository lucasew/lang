package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmallLangs(t *testing.T) {
	// Romanian/Slovak/Slovenian/Greek/Tagalog/Breton/Asturian/Belarusian/Danish/Esperanto promoted to dedicated Lang types (not SmallLang); remaining set ≥ 7.
	require.GreaterOrEqual(t, len(AllSmallLangs()), 7)
	require.Equal(t, "Ukrainian", Ukrainian.GetName())
}

func TestSmallLang_DefaultSpellingRuleIDs_MatchJava(t *testing.T) {
	// Faithful createDefaultSpellingRule / registered speller getId only.
	// Slovak speller id: SlovakLang.GetDefaultSpellingRuleID (slovak_test.go)
	// Danish speller id: DanishLang.GetDefaultSpellingRuleID (danish_test.go)
	require.Equal(t, "HUNSPELL_RULE", Swedish.GetDefaultSpellingRuleID())
	// Esperanto speller id: EsperantoLang.GetDefaultSpellingRuleID (esperanto_test.go)
	require.Equal(t, "HUNSPELL_RULE", Galician.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_RULE", Khmer.GetDefaultSpellingRuleID())
	require.Equal(t, "HUNSPELL_NO_SUGGEST_RULE", Icelandic.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", Irish.GetDefaultSpellingRuleID())
	// Breton speller id: BretonLang.GetDefaultSpellingRuleID (breton_test.go)
	require.Equal(t, "MORFOLOGIK_RULE_CRH_UA", CrimeanTatar.GetDefaultSpellingRuleID())
	// Slovenian speller id: SlovenianLang.GetDefaultSpellingRuleID (slovenian_test.go)
	require.Equal(t, "MORFOLOGIK_RULE_LT_LT", Lithuanian.GetDefaultSpellingRuleID())
	require.Equal(t, "MORFOLOGIK_RULE_ML_IN", Malayalam.GetDefaultSpellingRuleID())
	// No invent: languages without a Java default speller
	require.Empty(t, Japanese.GetDefaultSpellingRuleID())
	require.Empty(t, Chinese.GetDefaultSpellingRuleID())
	require.Empty(t, Persian.GetDefaultSpellingRuleID())
	require.Empty(t, Tamil.GetDefaultSpellingRuleID())
}

// Belarusian ignored chars / speller id: TestBelarusianMetadata in belarusian_test.go

func TestSmallLang_GetMaintainers(t *testing.T) {
	// Sample of Java getMaintainers ports
	require.Equal(t, "Leif-Jöran Olsson", Swedish.GetMaintainers()[0].Name)
	// Greek maintainers: TestGreekMetadata in greek_test.go
	// Esperanto maintainers: TestEsperantoMetadata in esperanto_test.go
	// Breton maintainers: TestBretonMetadata in breton_test.go
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

// Slovak GetRuleFileNames: TestSlovak_GetRuleFileNames in slovak_test.go

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
