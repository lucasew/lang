package languagetool

// Twin of FR JLanguageToolTest — Check inject + typography surface.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of JLanguageToolTest.testLanguageDependentFilter
func TestJLanguageTool_lang_fr_LanguageDependentFilter(t *testing.T) {
	lt := NewJLanguageTool("fr")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	require.Empty(t, lt.Check("Ceci est une phrase."))
	require.NotEmpty(t, lt.Check("Ceci est est une phrase."))
}

// Port of JLanguageToolTest.testMultitokenSpeller
func TestJLanguageTool_lang_fr_MultitokenSpeller(t *testing.T) {
	lt := NewJLanguageTool("fr")
	known := map[string]struct{}{"New": {}, "York": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, nil))
	require.Empty(t, lt.Check("New York"))
}

// Port of JLanguageToolTest.testMatchfiltering
func TestJLanguageTool_lang_fr_Matchfiltering(t *testing.T) {
	lt := NewJLanguageTool("fr")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{"Le": {}, "chat": {}, "dort": {}}, nil))
	// disable spell → only repeat would fire; no repeat here
	lt.DisableRule("SPELL")
	require.Empty(t, lt.Check("Le chat dort."))
	require.Equal(t, []string{"WORD_REPEAT_RULE"}, lt.GetAllActiveRuleIDs())
}

// Port of JLanguageToolTest.testQuotes
func TestJLanguageTool_lang_fr_Quotes(t *testing.T) {
	cfg := TypographyConfig{
		Enabled:            true,
		OpeningDoubleQuote: "«\u00a0",
		ClosingDoubleQuote: "\u00a0»",
		OpeningSingleQuote: "‘",
		ClosingSingleQuote: "’",
	}
	out := ToAdvancedTypography(`"C'est"`, cfg)
	require.Contains(t, out, "«")
	require.Contains(t, out, "»")
}

// Port of JLanguageToolTest.testMergingOfGrammarCorrections
func TestJLanguageTool_lang_fr_MergingOfGrammarCorrections(t *testing.T) {
	lt := NewJLanguageTool("fr")
	lt.AddRuleChecker("WORD_REPEAT_RULE", SimpleWordRepeatChecker("WORD_REPEAT_RULE"))
	// multi-sentence: repeat in second
	m := lt.Check("Phrase une. Phrase Phrase deux.")
	require.NotEmpty(t, m)
}
