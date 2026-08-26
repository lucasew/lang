package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterSuggestions_Prohibited(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddProhibitedWords("bad")
	got := r.FilterSuggestions([]string{"good", "bad", "ok"})
	require.Equal(t, []string{"good", "ok"}, got)
}

func TestFilterSuggestions_ProperNounS(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	// without IsProperNounFn fail-closed: keep "Michael s"
	got := r.FilterSuggestions([]string{"Michael s", "other"})
	require.Equal(t, []string{"Michael s", "other"}, got)

	r.IsProperNounFn = func(w string) bool { return w == "Michael" }
	got = r.FilterSuggestions([]string{"Michael s", "other"})
	// Java add(0,Name) then add(0,Name's) → [Name's, Name, other]
	require.Equal(t, []string{"Michael's", "Michael", "other"}, got)
}

func TestFilterSuggestions_Dupes(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	got := r.FilterSuggestions([]string{"a", "b", "a"})
	require.Equal(t, []string{"a", "b"}, got)
}

func TestFilterSuggestions_NoSuggestHook(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.FilterNoSuggestWordsFn = func(s []string) []string {
		var out []string
		for _, x := range s {
			if x != "blocked" {
				out = append(out, x)
			}
		}
		return out
	}
	got := r.FilterSuggestions([]string{"ok", "blocked"})
	require.Equal(t, []string{"ok"}, got)
}

func TestFilterSuggestions_TagPOS_NNP(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.TagPOS = func(w string) []string {
		if w == "Paris" {
			return []string{"NNP"}
		}
		return nil
	}
	got := r.FilterSuggestions([]string{"Paris s"})
	require.Equal(t, []string{"Paris's", "Paris"}, got)
}
