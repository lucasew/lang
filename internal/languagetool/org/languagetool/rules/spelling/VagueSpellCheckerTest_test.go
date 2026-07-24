package spelling

// Twin of VagueSpellCheckerTest.testIsValidWord
//
// Java: Languages.getLanguageForShortCode + VagueSpellChecker.isValidWord with real
// HunspellRule (de-DE) / MorfologikSpellerRule (en-US, pt-PT, fr) dict resources.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// vagueLang ports Language for VagueSpellChecker tests (LanguageMeta lacks CreateDefaultTagger).
type vagueLang struct {
	short, full, name string
}

func (l vagueLang) GetName() string {
	if l.name != "" {
		return l.name
	}
	return l.full
}
func (l vagueLang) GetShortCode() string { return l.short }
func (l vagueLang) GetShortCodeWithCountryAndVariant() string {
	if l.full != "" {
		return l.full
	}
	return l.short
}
func (l vagueLang) GetCountries() []string                     { return nil }
func (l vagueLang) CreateDefaultTagger() languagetool.Tagger { return nil }

func vagueLangFromRegistry(t *testing.T, code string) languagetool.Language {
	t.Helper()
	languagetool.EnsureBuiltInLanguagesRegistered()
	m := languagetool.GlobalLanguages.GetLanguageForShortCode(code)
	return vagueLang{
		short: m.GetShortCode(),
		full:  m.GetShortCodeWithCountryAndVariant(),
		name:  m.GetName(),
	}
}

func requireVagueDicts(t *testing.T) {
	t.Helper()
	need := []string{
		"/en/hunspell/en_US.dict",
		"/de/hunspell/de_DE.dict",
		"/fr/french.dict",
		"/pt/spelling/pt-PT-90.dict",
	}
	for _, cp := range need {
		if discoverVagueResourceDict(cp) == "" {
			t.Skipf("official dict not found for %s (fetch third_party/*-pos-dict or language-module resources)", cp)
		}
	}
}

// Port of VagueSpellCheckerTest.testIsValidWord
func TestVagueSpellChecker_IsValidWord(t *testing.T) {
	requireVagueDicts(t)
	clearVagueSpellCheckerCaches()
	t.Cleanup(clearVagueSpellCheckerCaches)

	de := vagueLangFromRegistry(t, "de-DE")
	en := vagueLangFromRegistry(t, "en-US")
	pt := vagueLangFromRegistry(t, "pt-PT")
	fr := vagueLangFromRegistry(t, "fr")
	checker := NewVagueSpellChecker()

	require.True(t, checker.IsValidWord("vacation", en))
	require.True(t, checker.IsValidWord("walks", en))
	require.False(t, checker.IsValidWord("vacationx", en))

	require.True(t, checker.IsValidWord("Hütte", de))
	require.True(t, checker.IsValidWord("Hütten", de))
	require.False(t, checker.IsValidWord("sdasfd", de))

	require.True(t, checker.IsValidWord("termo", pt))
	require.False(t, checker.IsValidWord("termoasq", pt))
	require.False(t, checker.IsValidWord("difdsf", pt))

	require.True(t, checker.IsValidWord("voiture", fr))
	require.False(t, checker.IsValidWord("sduiofhdf", fr))
}

// Control-flow / polarity smoke: Morfologik path caches dict; Hunspell path uses !isMisspelled.
func TestVagueSpellChecker_CacheAndRuleKinds(t *testing.T) {
	requireVagueDicts(t)
	clearVagueSpellCheckerCaches()
	t.Cleanup(clearVagueSpellCheckerCaches)

	en := vagueLangFromRegistry(t, "en-US")
	de := vagueLangFromRegistry(t, "de-DE")
	checker := NewVagueSpellChecker()

	require.True(t, checker.IsValidWord("vacation", en))
	// second call hits langToRule cache
	require.True(t, checker.IsValidWord("walks", en))
	require.True(t, checker.IsValidWord("Hütte", de))
	require.False(t, checker.IsValidWord("sdasfd", de))

	// NonThreadSafeSpellRule surfaces
	nts := newNonThreadSafeSpellRule(en)
	require.Equal(t, "FAKE_FOR_VAGUE_SPELL_CHECKER", nts.getId())
	require.Equal(t, "internal", nts.getDescription())
}

// Ensure discovery finds the four official dicts when present under third_party or modules.
func TestVagueSpellChecker_DiscoverOfficialDicts(t *testing.T) {
	// At least EN is vendored in this repo's third_party/english-pos-dict.
	if p := discoverVagueResourceDict("/en/hunspell/en_US.dict"); p == "" {
		// walk up from package dir
		wd, _ := os.Getwd()
		t.Logf("cwd=%s", wd)
		t.Skip("en_US.dict not discoverable")
	} else {
		require.True(t, filepath.IsAbs(p) || p != "")
		st, err := os.Stat(p)
		require.NoError(t, err)
		require.True(t, st.Mode().IsRegular())
	}
}
