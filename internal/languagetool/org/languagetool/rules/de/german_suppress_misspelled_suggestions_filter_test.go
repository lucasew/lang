package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestGermanSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewGermanSuppressMisspelledSuggestionsFilter()
	// no dict: FilterDictIsMisspelled returns false (Java null SpellingCheckRule)
	kept, ok := f.FilterSuggestions([]string{"Haus", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"Haus", "xyz"}, kept)
}

func TestGermanSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewGermanSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"Haus", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"Haus"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestGermanSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewGermanSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"Haus", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Haus"}, out.GetSuggestedReplacements())

	// all misspelled + suppressMatch → drop
	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))

	// all misspelled + suppressMatch false → keep match with empty suggestions
	m3 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m3.SetSuggestedReplacements([]string{"xyz"})
	out = f.AcceptRuleMatch(m3, map[string]string{"suppressMatch": "false"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestGermanSuppressMisspelledSuggestionsFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewGermanSuppressMisspelledSuggestionsFilter()
	// Haus known, nonsense misspelled
	kept, ok := f.FilterSuggestions([]string{"Haus", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "Haus")
	require.NotContains(t, kept, "xyzzyqqq")
	// all misspelled + suppress → drop match
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"xyzzyqqq"})
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestGermanSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.GermanSuppressMisspelledSuggestionsFilter"))
}
