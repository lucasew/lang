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
