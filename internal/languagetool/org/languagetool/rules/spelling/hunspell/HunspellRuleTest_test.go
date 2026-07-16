package hunspell

// Twin of HunspellRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestHunspellRule_HighConfidenceSuggestion(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello", "world"})
	dict.SetSuggestions("helo", []string{"hello"})
	r := NewHunspellRule("en", dict)
	require.True(t, r.IsMisspelledWord("helo"))
	require.False(t, r.IsMisspelledWord("hello"))
	require.Equal(t, []string{"hello"}, r.Suggest("helo"))
}

func TestHunspellRule_RuleWithGerman(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"Haus", "Hund", "und"})
	dict.SetSuggestions("Huas", []string{"Haus"})
	r := NewHunspellRule("de", dict)
	sent := languagetool.AnalyzePlain("Haus und Huas")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, []string{"Haus"}, matches[0].GetSuggestedReplacements())
}

func TestHunspellRule_MultitokensWithSepllerRule(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"foo", "bar"})
	r := NewHunspellRule("en", dict)
	sent := languagetool.AnalyzePlain("foo bar baz")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1) // baz
}

func TestHunspellRule_RuleWithWrongSplit(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"wrong", "split"})
	r := NewHunspellRule("en", dict)
	require.True(t, r.IsMisspelledWord("wrongs"))
}

func TestHunspellRule_RuleWithAustrianGerman(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"Jänner"})
	r := NewHunspellRule("de-AT", dict)
	require.False(t, r.IsMisspelledWord("Jänner"))
}

func TestHunspellRule_RuleWithSwissGerman(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"ss", "Strasse"})
	r := NewHunspellRule("de-CH", dict)
	require.False(t, r.IsMisspelledWord("Strasse"))
}

func TestHunspellRule_Performance(t *testing.T) {
	t.Skip("Java @Ignore")
}

func TestHunspellRule_CompoundAwareRulePerformance(t *testing.T) {
	t.Skip("Java @Ignore")
}
