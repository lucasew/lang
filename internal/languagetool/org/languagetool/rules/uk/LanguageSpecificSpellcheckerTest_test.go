package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/LanguageSpecificSpellcheckerTest.java
// Java: extends SpellcheckerTest; testRules() → runTests("uk").
// Full multi-language SpellcheckerTest harness deferred — ID + official dict smoke.
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

	if morfologik.DiscoverLanguageDict(UkrainianSpellerDict) == "" {
		t.Skip("uk_UA.dict not discoverable")
	}
	require.NotNil(t, r.Speller1)
	// known good word should not be misspelled by multi
	require.False(t, r.IsMisspelled("тест"))
	require.True(t, r.IsMisspelled("слввоxyzzyqqq"))
}
