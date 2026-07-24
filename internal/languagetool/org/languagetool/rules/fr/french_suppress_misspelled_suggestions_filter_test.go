package fr

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFrenchSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"maison", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"maison", "xyz"}, kept)
}

func TestFrenchSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"maison", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"maison"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestFrenchSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"maison", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"maison"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestFrenchSuppressMisspelledSuggestionsFilter_WithFRDict(t *testing.T) {
	ClearFrenchFilterSpeller()
	t.Cleanup(ClearFrenchFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/fr/src/main/resources/org/languagetool/resource/fr/french.dict"),
		filepath.Join(root, "third_party/fr/french.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireFrenchFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("french.dict not in tree (Java French default spelling)")
	}
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"maison", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "maison")
	require.NotContains(t, kept, "xyzzyqqq")
}

func TestFrenchSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.FrenchSuppressMisspelledSuggestionsFilter"))
}
