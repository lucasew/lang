package morfologik

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetWeightedSuggestions_MaxMatchLength(t *testing.T) {
	// Java: word.length() > StringMatcher.MAX_MATCH_LENGTH (250) → empty
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("ok")
	long := strings.Repeat("a", stringMatcherMaxMatchLength+1)
	require.Empty(t, sp.GetWeightedSuggestions(long))
	// boundary: length 250 still allowed through length gate (may still be empty of sugs)
	boundary := strings.Repeat("a", stringMatcherMaxMatchLength)
	_ = sp.GetWeightedSuggestions(boundary) // must not panic
}

func TestGetWeightedSuggestions_SkipFindReplWhenLong(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	// classic short misspelling still works
	require.Contains(t, sp.FindReplacements("recieve"), "receive")
	// word length >= 50: Java skips findReplacementCandidates; only run-ons possible
	longMiss := strings.Repeat("x", morfologikFindReplMaxLen) // 50 x's, misspelled
	sugs := sp.GetWeightedSuggestions(longMiss)
	// must not return normal edit-distance neighbors (would be expensive / invent)
	for _, s := range sugs {
		// run-ons have a space; pure edit sugs would not for all-x input
		if !strings.Contains(s.Word, " ") {
			t.Fatalf("len>=50 should not run findRepl; got non-runon %q among %v", s.Word, sugs)
		}
	}
}

func TestApplyCaseToWeighted_AllUpper(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.ConvertCase = true
	in := []WeightedSuggestion{NewWeightedSuggestion("receive", 10)}
	out := applyCaseToWeighted(sp, "RECIEVE", in)
	require.Equal(t, "RECEIVE", out[0].Word)
}

func TestApplyCaseToWeighted_TitleCase(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.ConvertCase = true
	in := []WeightedSuggestion{NewWeightedSuggestion("receive", 10)}
	out := applyCaseToWeighted(sp, "Recieve", in)
	require.Equal(t, "Receive", out[0].Word)
}

// StringTools.uppercaseFirstChar skips leading non-letter/digit (quotes/parens).
func TestApplyCaseToWeighted_LeadingQuote(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.ConvertCase = true
	in := []WeightedSuggestion{NewWeightedSuggestion("\"hello\"", 10)}
	out := applyCaseToWeighted(sp, "Xyz", in) // title arm
	// changeFirstCharCase: pos after ", uppercases h → "Hello"
	require.Equal(t, "\"Hello\"", out[0].Word)
}

func TestApplyCaseToWeighted_MixedCaseSuggestionUnchanged(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.ConvertCase = true
	// mixedCase suggestion must stay when input is title case (Java StringTools.isMixedCase)
	in := []WeightedSuggestion{NewWeightedSuggestion("iPhone", 10)}
	out := applyCaseToWeighted(sp, "Iphone", in)
	require.Equal(t, "iPhone", out[0].Word)
}
