package languagetool

// Twin of JLanguageToolTest homepage demos — Check inject (full EN grammar deferred).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_DemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	// soft spelling for "tot" / "he" style is deferred; a/an catches "a error"
	matches := lt.Check("A sentence with a error in the Hitchhiker's Guide tot he Galaxy")
	require.NotEmpty(t, matches)
	var found bool
	for _, m := range matches {
		if m.RuleID == "EN_A_VS_AN" {
			found = true
			require.Contains(t, m.Suggestions, "an")
		}
	}
	require.True(t, found, "expected a→an for 'a error'")
}

func TestJLanguageTool_SpellCheckerDemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{
		"A": {}, "a": {}, "error": {}, "spelling": {},
	}
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", SimpleMapSpellerChecker("MORFOLOGIK_RULE_EN_US", known, map[string][]string{
		"speling": {"spelling"},
	}))
	matches := lt.Check("A speling error")
	require.NotEmpty(t, matches)
	require.Equal(t, "speling", "A speling error"[2:9])
	// match covers speling
	require.Equal(t, []string{"spelling"}, matches[0].Suggestions)
	fixed := CorrectTextFromLocalMatches("A speling error", matches)
	require.Equal(t, "A spelling error", fixed)
}

func TestJLanguageTool_SpellCheckerDemoCodeForHomepageWithAddedWords(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{"LanguageTool": {}}
	lt.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", known, nil))
	// accepted after "adding" to dict
	require.Empty(t, lt.Check("LanguageTool"))
	// without word → misspelled
	lt2 := NewJLanguageTool("en-US")
	lt2.AddRuleChecker("SPELL", SimpleMapSpellerChecker("SPELL", map[string]struct{}{}, nil))
	require.NotEmpty(t, lt2.Check("LanguageTool"))
}
