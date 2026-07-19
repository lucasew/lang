package en

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestEnglishSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"house", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"house", "xyz"}, kept)
}

func TestEnglishSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"house", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"house"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestEnglishSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"house", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"house"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestEnglishSuppressMisspelledSuggestionsFilter_WithENDict(t *testing.T) {
	ClearEnglishFilterSpeller()
	t.Cleanup(ClearEnglishFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	if !WireEnglishFilterSpeller(dict) {
		t.Skipf("en_US.dict not openable: %s", dict)
	}
	f := NewEnglishSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"house", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "house")
	require.NotContains(t, kept, "xyzzyqqq")
}

func TestEnglishSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter")
	require.NotNil(t, f)
}
