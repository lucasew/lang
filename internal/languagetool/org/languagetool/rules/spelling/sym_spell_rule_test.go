package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	symimpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
	"github.com/stretchr/testify/require"
)

func TestSymSpellRule(t *testing.T) {
	r := NewSymSpellRule("SYMSPELL_RULE", "en")
	r.AddWords("hello", "world")
	require.True(t, r.isMisspelled("helo"))
	require.False(t, r.isMisspelled("hello"))
	sent := languagetool.AnalyzePlain("hello helo")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

// Twin: Match uses implementation.SymSpell.lookup when Speller is set.
func TestSymSpellRule_WithSpellerLookup(t *testing.T) {
	sp := symimpl.DefaultSymSpell()
	require.True(t, sp.CreateDictionaryEntry("hello", 10, nil))
	require.True(t, sp.CreateDictionaryEntry("world", 5, nil))
	r := NewSymSpellRule(SymSpellRuleID, "en")
	r.SetSpeller(sp)
	// EditDistance 3 but speller max is 2 — clamped
	r.EditDistance = 3
	require.False(t, r.isMisspelled("hello"))
	require.True(t, r.isMisspelled("helo"))
	sugs := r.Suggestions("helo")
	require.Contains(t, sugs, "hello")
	sent := languagetool.AnalyzePlain("helo world")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	// first match is helo
	require.Contains(t, matches[0].GetSuggestedReplacements(), "hello")
}

// Twin: filterCandidates drops prohibited.
func TestSymSpellRule_FilterProhibited(t *testing.T) {
	r := NewSymSpellRule(SymSpellRuleID, "en")
	r.AddWords("hello", "hallo")
	r.Prohibited["hallo"] = struct{}{}
	sugs := r.Suggestions("helo")
	require.Contains(t, sugs, "hello")
	require.NotContains(t, sugs, "hallo")
}
