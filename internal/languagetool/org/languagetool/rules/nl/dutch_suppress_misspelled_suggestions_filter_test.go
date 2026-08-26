package nl

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDutchSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearDutchFilterSpeller()
	f := NewDutchSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"huis", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"huis", "xyz"}, kept)
}

func TestDutchSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearDutchFilterSpeller()
	f := NewDutchSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"huis", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"huis"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestDutchSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearDutchFilterSpeller()
	f := NewDutchSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"huis", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"huis"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestDutchSuppressMisspelledSuggestionsFilter_WithNLDict(t *testing.T) {
	ClearDutchFilterSpeller()
	t.Cleanup(ClearDutchFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/nl/src/main/resources/org/languagetool/resource/nl/spelling/nl_NL.dict"),
		filepath.Join(root, "third_party/nl/spelling/nl_NL.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireDutchFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("nl_NL.dict not in tree (Java Dutch default spelling)")
	}
	f := NewDutchSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"huis", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "huis")
	require.NotContains(t, kept, "xyzzyqqq")
}

func TestDutchSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.nl.DutchSuppressMisspelledSuggestionsFilter"))
}
