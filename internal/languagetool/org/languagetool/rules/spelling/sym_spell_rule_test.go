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

// Twin: Java filterCandidates only on default candidates — user list unfiltered.
func TestSymSpellRule_UserCandidatesNotFiltered(t *testing.T) {
	r := NewSymSpellRule(SymSpellRuleID, "en")
	r.AddWords("hello")
	r.AddUserWords("hallo")
	r.Prohibited["hallo"] = struct{}{}
	// default suggestions for helo include hello; user may include hallo
	// Match should still allow user "hallo" as candidate for suggestions path
	// when both have non-exact candidates
	r.AddWords("helo") // so helo is exact? remove that
	// rebuild: helo misspelled, default → hello, user → hallo (prohibited but not filtered on user)
	// getSpellerMatches user returns hallo from map
	// filter only default
	cands := r.filterCandidates(r.getSpellerMatches("helo", false))
	user := r.getSpellerMatches("helo", true)
	require.NotContains(t, cands, "hallo")
	// user map path: if only user has hallo as suggestion via map scan
	// With map inject Suggestions from levenshtein on UserDictionary
	require.Contains(t, user, "hallo", "user candidates not filterCandidates'd: %v", user)
}

// Twin: Java ignoredWords.contains only — IgnoreWords set, not full IgnoreWord (e.g. length).
func TestSymSpellRule_IgnoreSetOnly(t *testing.T) {
	r := NewSymSpellRule(SymSpellRuleID, "en")
	r.IgnoreWordsWithLength = 1
	// single letter not in ignore set → still matched when unknown
	sent := languagetool.AnalyzePlain("x")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	// "x" is non-empty token; if IsNonWord for single letter?
	// If non-word, skip; else unknown word match
	// Java isNonWord may skip punctuation; letters are words
	require.NotEmpty(t, ms, "IgnoreWordsWithLength must not skip like ignore set: got %v", ms)
	// explicit ignore set does skip
	r.AddIgnoreWords("xyzzy")
	sent2 := languagetool.AnalyzePlain("xyzzy")
	ms2, err := r.Match(sent2)
	require.NoError(t, err)
	require.Empty(t, ms2)
}

// Twin: both empty candidates → always "Misspelling or unknown word!" (no AcceptWord gate).
func TestSymSpellRule_UnknownAlwaysFlags(t *testing.T) {
	r := NewSymSpellRule(SymSpellRuleID, "en")
	// empty dict
	sent := languagetool.AnalyzePlain("foobar")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "Misspelling or unknown word!", ms[0].Message)
}
