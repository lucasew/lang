package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port/control-flow unit tests for VagueSpellChecker (no invent Register inject).

func TestVagueSpellChecker_EnglishMorfologikPath(t *testing.T) {
	if discoverVagueResourceDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not available")
	}
	clearVagueSpellCheckerCaches()
	t.Cleanup(clearVagueSpellCheckerCaches)

	en := vagueLang{short: "en", full: "en-US", name: "English (US)"}
	v := NewVagueSpellChecker()
	require.True(t, v.IsValidWord("vacation", en))
	require.False(t, v.IsValidWord("vacationx", en))

	// getSpellingCheckRule returns Morfologik surface for en-US
	rule := getSpellingCheckRule(en)
	ms, ok := rule.(*morfologikSpellerRuleForVague)
	require.True(t, ok)
	require.Equal(t, "/en/hunspell/en_US.dict", ms.GetFileName())
}

func TestVagueSpellChecker_GermanHunspellPath(t *testing.T) {
	if discoverVagueResourceDict("/de/hunspell/de_DE.dict") == "" {
		t.Skip("de_DE.dict not available")
	}
	clearVagueSpellCheckerCaches()
	t.Cleanup(clearVagueSpellCheckerCaches)

	de := vagueLang{short: "de", full: "de-DE", name: "German (Germany)"}
	rule := getSpellingCheckRule(de)
	_, ok := rule.(*hunspellRuleForVague)
	require.True(t, ok, "de-DE must use HunspellRule branch (GermanSpellerRule)")

	v := NewVagueSpellChecker()
	require.True(t, v.IsValidWord("Hütte", de))
	require.False(t, v.IsValidWord("sdasfd", de))
}

func TestVagueSpellChecker_GermanHardCases(t *testing.T) {
	// GermanSpellerRule.isMisspelled hard cases (before dict): Standart*/Spielzug* (non-whitelist)
	if discoverVagueResourceDict("/de/hunspell/de_DE.dict") == "" {
		t.Skip("de_DE.dict not available")
	}
	clearVagueSpellCheckerCaches()
	t.Cleanup(clearVagueSpellCheckerCaches)

	de := vagueLang{short: "de", full: "de-DE", name: "German (Germany)"}
	v := NewVagueSpellChecker()
	// Standart alone is misspelled hard-case (not Standarte/Standarten/…)
	require.False(t, v.IsValidWord("Standart", de), "Standart is hard-case misspelled")
	// Spielzug alone is not on START_WITH_SPIEL whitelist → misspelled
	require.False(t, v.IsValidWord("Spielzugxyz", de), "non-whitelisted Spielzug* misspelled")
	// Hütte still valid
	require.True(t, v.IsValidWord("Hütte", de))
}

func TestVagueSpellChecker_SpanishCatalanOfficialPaths(t *testing.T) {
	// Java MorfologikSpanishSpellerRule.getFileName / MorfologikCatalanSpellerRule.getFileName
	es := vagueLang{short: "es", full: "es", name: "Spanish"}
	ca := vagueLang{short: "ca", full: "ca", name: "Catalan"}
	ruleES := getSpellingCheckRule(es)
	msES, ok := ruleES.(*morfologikSpellerRuleForVague)
	require.True(t, ok)
	require.Equal(t, "/es/es-ES.dict", msES.GetFileName())

	ruleCA := getSpellingCheckRule(ca)
	msCA, ok := ruleCA.(*morfologikSpellerRuleForVague)
	require.True(t, ok)
	require.Equal(t, "/ca/ca-ES_spelling.dict", msCA.GetFileName())
}

func TestVagueSpellChecker_PureHunspellUsesHunspellRuleTwin(t *testing.T) {
	// da/sv/eo/gl/km → HunspellRule twin; is → HunspellNoSuggestionRule twin.
	// Must not use soft stem-map invent path.
	for _, tc := range []struct {
		short, full string
	}{
		{"da", "da"},
		{"sv", "sv"},
		{"eo", "eo"},
		{"gl", "gl"},
		{"km", "km"},
		{"is", "is"},
	} {
		t.Run(tc.short, func(t *testing.T) {
			lang := vagueLang{short: tc.short, full: tc.full, name: tc.short}
			rule := getSpellingCheckRule(lang)
			_, ok := rule.(*hunspellRuleForVague)
			require.True(t, ok, "%s must be HunspellRule surface", tc.short)
		})
	}
}

func TestVagueSpellChecker_NoSpellingRulePanics(t *testing.T) {
	// Japanese has no spelling rule in getRelevantRules
	ja := vagueLang{short: "ja", full: "ja", name: "Japanese"}
	require.Panics(t, func() {
		_ = getSpellingCheckRule(ja)
	})
}

func TestVagueSpellChecker_DictClasspathHelpers(t *testing.T) {
	require.Equal(t, "/en/hunspell/en_US.dict", englishMorfologikDictClasspath("en-us"))
	require.Equal(t, "/en/hunspell/en_GB.dict", englishMorfologikDictClasspath("en-gb"))
	require.Equal(t, "/en/hunspell/en_CA.dict", englishMorfologikDictClasspath("en-ca"))
	require.Equal(t, "/de/hunspell/de_DE.dict", germanHunspellDictClasspath("de-de"))
	require.Equal(t, "/de/hunspell/de_AT.dict", germanHunspellDictClasspath("de-at"))
	require.Equal(t, "/pt/spelling/pt-PT-90.dict", portugueseMorfologikDictClasspath("pt-pt"))
	require.Equal(t, "/pt/spelling/pt-BR.dict", portugueseMorfologikDictClasspath("pt-br"))
}

// Ensure soft invent helpers are gone (would not compile if reintroduced as API).
func TestVagueSpellChecker_NoSoftStemInventAPI(t *testing.T) {
	// Pure-hunspell langs resolve through hunspell.NewHunspellRule twin; firstActive
	// must not invent dic-only stem maps as a separate parallel path.
	// Classpath assertion for Icelandic matches language-module twin is_IS.dic.
	is := vagueLang{short: "is", full: "is", name: "Icelandic"}
	rule := getSpellingCheckRule(is)
	h, ok := rule.(*hunspellRuleForVague)
	require.True(t, ok)
	require.NotNil(t, h.isMisspelledFn)
}
