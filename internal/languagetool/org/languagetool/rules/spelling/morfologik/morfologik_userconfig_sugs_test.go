package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGetSpellingSuggestions(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	sugs := r.GetSpellingSuggestions("recieve")
	require.Contains(t, sugs, "receive")
	require.Empty(t, r.GetSpellingSuggestions("receive"), "known word → no match sugs")
	require.Empty(t, r.GetSpellingSuggestions(""))
}

func TestMatch_SuggestionsDisabled(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	uc := languagetool.NewUserConfig()
	uc.SuggestionsEnabled = false
	r.SetUserConfig(uc)
	ms, err := r.Match(languagetool.AnalyzePlain("recieve"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// Java: match still emitted, no dict suggestions
	require.Empty(t, ms[0].GetSuggestedReplacements())
}

func TestMatch_MaxSpellingSuggestionsLimit(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	sp.AddWord("the")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	uc := languagetool.NewUserConfig()
	uc.MaxSpellingSuggestions = 1
	r.SetUserConfig(uc)
	// Two misspellings: first still gets sugs (soFar=0 <= 1); second hits limit (soFar=1... wait 1<=1 still allows)
	// Java: soFar <= max → when soFar is 1 and max is 1, still allows. soFar=2 and max=1 blocks.
	// Three misspellings to exceed after first two.
	ms, err := r.Match(languagetool.AnalyzePlain("recieve recieve recieve"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ms), 2)
	// First match (soFar=0) has real sugs
	require.NotEmpty(t, ms[0].GetSuggestedReplacements())
	require.NotContains(t, ms[0].GetSuggestedReplacements(), tooManyErrorsMsg)
	// When soFar > max: third match gets too_many_errors (if 3 matches)
	if len(ms) >= 3 {
		// soFar=2 for third: 2 <= 1 is false
		require.Equal(t, []string{tooManyErrorsMsg}, ms[2].GetSuggestedReplacements())
	}
}
