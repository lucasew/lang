package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestCompoundAwareHunspellSuggest(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"well", "known", "wellknown"})
	morfo := morfologik.NewMorfologikSpeller("en", 1)
	morfo.AddWord("well")
	morfo.AddWord("known")
	morfo.Suggestions["wel"] = []string{"well"}
	multi := morfologik.NewMorfologikMultiSpeller(morfo)
	r := NewCompoundAwareHunspellRule("en", dict, nil, multi)
	// compound split of well-known
	sug := r.Suggest("well-known")
	require.NotEmpty(t, sug)
	// misspelled with morfo suggestion
	dict2 := NewMapHunspellDictionary([]string{"well", "known"})
	r2 := NewCompoundAwareHunspellRule("en", dict2, nil, multi)
	require.NotEmpty(t, r2.Suggest("wel-known"))
}

// Twin: sortSuggestionByQuality prefers run-on (space-removed == misspelling).
func TestCompoundAware_SortSuggestionByQuality_RunOn(t *testing.T) {
	r := NewCompoundAwareHunspellRule("en", NewMapHunspellDictionary(nil), nil, nil)
	got := r.sortSuggestionByQuality("thankyou", []string{"thank", "thank you", "thanks"})
	require.Equal(t, "thank you", got[0], "run-on preferred: %v", got)
	// single-letter split-off is NOT preferred
	got = r.sortSuggestionByQuality("athank", []string{"a thank", "thank"})
	require.Equal(t, "a thank", got[0]) // not reordered to front when single letter
	require.Equal(t, []string{"a thank", "thank"}, got)
}

// Twin: StringUtils.split on ' ' only — tab is not a token boundary for single-letter check.
func TestHasSingleLetterToken_ASCIISpaceOnly(t *testing.T) {
	require.True(t, hasSingleLetterToken("a thank"))
	require.False(t, hasSingleLetterToken("thank you"))
	// tab-joined single letter is one token under StringUtils.split(' ')
	require.False(t, hasSingleLetterToken("a\tthank"))
	// UTF-16 length-1 for multi-byte letter still counts as single letter token
	require.True(t, hasSingleLetterToken("é word"))
}

// Twin: interleave order noSplit, upper(lc), simple.
func TestInterleaveThree(t *testing.T) {
	got := interleaveThree(
		[]string{"a1", "a2"},
		[]string{"b1"},
		[]string{"c1", "c2", "c3"},
	)
	require.Equal(t, []string{"a1", "b1", "c1", "a2", "c2", "c3"}, got)
}

// SuggestFn is wired so Match/calcSuggestions use compound-aware getSuggestions.
func TestCompoundAware_SuggestFnUsedByMatch(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello"})
	morfo := morfologik.NewMorfologikSpeller("en", 1)
	morfo.Suggestions["helo"] = []string{"hello"}
	multi := morfologik.NewMorfologikMultiSpeller(morfo)
	r := NewCompoundAwareHunspellRule("en", dict, nil, multi)
	require.NotNil(t, r.HunspellRule.SuggestFn)

	// Match uses Suggest via SuggestFn → morfo suggestions appear
	ms, err := r.Match(languagetool.AnalyzePlain("helo"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "hello")
}

// handleWordEndPunctuation: "informationnen." → morfo on base + reappend "."
func TestHandleWordEndPunctuation(t *testing.T) {
	morfo := morfologik.NewMorfologikSpeller("en", 1)
	morfo.Suggestions["informationnen"] = []string{"information"}
	multi := morfologik.NewMorfologikMultiSpeller(morfo)
	var noSplit []string
	handleWordEndPunctuation(".", "informationnen.", &noSplit, multi)
	require.Contains(t, noSplit, "information.")
}

// getCorrectWords filters via Hunspell.spell per token.
func TestGetCorrectWords(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"due", "to"})
	r := NewCompoundAwareHunspellRule("en", dict, nil, nil)
	got := r.getCorrectWords([]string{"due", "due to", "due xyz"})
	require.Contains(t, got, "due")
	require.Contains(t, got, "due to")
	require.NotContains(t, got, "due xyz")
}

// GetCandidatesFromParts rebuilds when a middle part is misspelled.
func TestGetCandidatesFromParts(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"well", "known", "wellknown"})
	morfo := morfologik.NewMorfologikSpeller("en", 1)
	morfo.Suggestions["knon"] = []string{"known"}
	// also need Spell for parts
	morfo.AddWord("well")
	morfo.AddWord("known")
	multi := morfologik.NewMorfologikMultiSpeller(morfo)
	r := NewCompoundAwareHunspellRule("en", dict, nil, multi)
	// well is spelled, knon is not → rebuild well+known
	cands := r.GetCandidatesFromParts([]string{"well", "knon"})
	require.Contains(t, cands, "wellknown")
}

// SpellingFilePaths matches Java getSpellingFilePaths.
func TestSpellingFilePaths(t *testing.T) {
	require.Equal(t, []string{
		"/de/hunspell/spelling.txt",
		"/de/hunspell/spelling_custom.txt",
		"/de/multitoken-suggest.txt",
		"spelling_global.txt",
	}, SpellingFilePaths("de"))
}

// tokenizeTextHun splits on non-letters (Java NON_ALPHABETIC).
func TestTokenizeTextHun(t *testing.T) {
	require.Equal(t, []string{"due", "to"}, tokenizeTextHun("due to"))
	require.Equal(t, []string{"well", "known"}, tokenizeTextHun("well-known"))
	require.Empty(t, tokenizeTextHun(""))
}
