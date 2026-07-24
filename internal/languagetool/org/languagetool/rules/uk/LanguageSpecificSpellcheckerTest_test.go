package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/LanguageSpecificSpellcheckerTest.java
// Full SpellcheckerTest dict deferred — analyze + speller ID smoke.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificSpellcheckerTest.testRules
func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	lt := languagetool.NewJLanguageTool("uk")
	require.NotEmpty(t, lt.Analyze("тест"))

	r := NewMorfologikUkrainianSpellerRule()
	require.Equal(t, MorfologikUkrainianSpellerRuleID, r.GetID())
	require.Equal(t, UkrainianSpellerDict, r.GetFileName())

	// inject: known word OK (tagged → ignoreToken hasGoodTag), misspelling flagged
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 1)
	sp.AddWord("тест")
	sp.AddWord("мова")
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	// Java: tagged Ukrainian words skip spellcheck via hasGoodTag
	m, err := ukMatchTagged(r, "тест мова")
	require.NoError(t, err)
	require.Empty(t, m)
	// non-Ukrainian letters ignored (not flagged); use untagged misspell
	m, err = r.Match(languagetool.AnalyzePlain("слвво"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
}
