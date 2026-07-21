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
	// Java HunspellRule wrong-split: "thanky ou" → "thank you"
	dict := NewMapHunspellDictionary([]string{"thank", "you", "wrong", "split"})
	r := NewHunspellRule("en", dict)
	require.True(t, r.IsMisspelledWord("thanky"))
	require.True(t, r.IsMisspelledWord("ou"))
	require.False(t, r.IsMisspelledWord("thank"))
	require.False(t, r.IsMisspelledWord("you"))

	sent := languagetool.AnalyzePlain("thanky ou")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	// First match should be wrong-split covering both tokens
	require.Contains(t, matches[0].GetSuggestedReplacements(), "thank you")
	// Span from start of thanky through ou (UTF-16 positions from AnalyzePlain)
	require.Equal(t, 0, matches[0].GetFromPos())
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
