package languagetool

// Twin of JLanguageToolTest homepage demos — Check inject (full EN grammar deferred).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_DemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	lt.RegisterDemoEnglishCheckers(map[string]struct{}{
		"A": {}, "sentence": {}, "with": {}, "error": {}, "in": {}, "the": {},
		"Hitchhiker": {}, "Guide": {}, "Galaxy": {}, "to": {},
	}, nil)
	src := "A sentence with a error in the Hitchhiker's Guide tot he Galaxy"
	matches := lt.Check(src)
	require.NotEmpty(t, matches)
	ids := map[string]bool{}
	for _, m := range matches {
		ids[m.RuleID] = true
	}
	require.True(t, ids["EN_A_VS_AN"], "expected a→an for 'a error'")
	require.True(t, ids["PHRASE_REPLACE"], "expected tot he → to the")
	// apply all suggestions iteratively
	fixed := src
	for {
		ms := lt.Check(fixed)
		if len(ms) == 0 {
			break
		}
		// only apply first match with suggestions per pass
		applied := false
		for _, m := range ms {
			if len(m.Suggestions) > 0 {
				fixed = CorrectTextFromLocalMatches(fixed, []LocalMatch{m})
				applied = true
				break
			}
		}
		if !applied {
			break
		}
	}
	require.Contains(t, fixed, "an error")
	require.Contains(t, fixed, "to the")
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
