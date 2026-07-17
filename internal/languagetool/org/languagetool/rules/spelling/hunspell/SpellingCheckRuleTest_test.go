package hunspell

// Twin of SpellingCheckRuleTest (DE hunspell module) — inject Map dictionary greens.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

// Port of SpellingCheckRuleTest.testIgnoreSuggestionsWithHunspell
func TestSpellingCheckRule_IgnoreSuggestionsWithHunspell(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"Haus", "Baum"})
	r := NewHunspellRule("de", dict)
	require.False(t, r.IsMisspelledWord("Haus"))
	require.True(t, r.IsMisspelledWord("Huas"))
}

// Port of SpellingCheckRuleTest.testIgnorePhrases
func TestSpellingCheckRule_IgnorePhrases(t *testing.T) {
	r := spelling.NewSpellingCheckRule("SPELL", "spelling", "de")
	r.IsMisspelled = func(word string) bool { return word == "xyz" }
	r.AddIgnoreWords("LanguageTool", "xyz")
	require.True(t, r.AcceptWord("xyz"))
	require.True(t, r.AcceptWord("LanguageTool"))
}

// Port of SpellingCheckRuleTest.testMultitokenSpelling
func TestSpellingCheckRule_MultitokenSpelling(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"New", "York", "City"})
	r := NewHunspellRule("en", dict)
	sent := languagetool.AnalyzePlain("New York City")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)
}

// Port of SpellingCheckRuleTest.testProhibitedWordFollowedByDot
func TestSpellingCheckRule_ProhibitedWordFollowedByDot(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"ok"})
	r := NewHunspellRule("de", dict)
	// "Huas." — word token is still misspelled; period is separate
	sent := languagetool.AnalyzePlain("Huas.")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
}
