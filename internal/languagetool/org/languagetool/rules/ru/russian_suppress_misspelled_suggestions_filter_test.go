package ru

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRussianSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearRussianFilterSpeller()
	f := NewRussianSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"дом", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"дом", "xyz"}, kept)
}

func TestRussianSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearRussianFilterSpeller()
	f := NewRussianSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"дом", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"дом"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestRussianSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearRussianFilterSpeller()
	f := NewRussianSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"дом", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"дом"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestRussianSuppressMisspelledSuggestionsFilter_WithRUDict(t *testing.T) {
	ClearRussianFilterSpeller()
	t.Cleanup(ClearRussianFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java MorfologikRussianSpellerRule: /ru/hunspell/ru_RU.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/ru/src/main/resources/org/languagetool/resource/ru/hunspell/ru_RU.dict"),
		filepath.Join(root, "third_party/ru/hunspell/ru_RU.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireRussianFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("ru_RU.dict not in tree (Java MorfologikRussianSpellerRule)")
	}
	f := NewRussianSuppressMisspelledSuggestionsFilter()
	// junk must be misspelled
	kept, ok := f.FilterSuggestions([]string{"xyzzyqqq"}, true)
	require.False(t, ok)
	require.Empty(t, kept)
	// known Russian word if in dict
	if !FilterDictIsMisspelled("дом") {
		kept2, ok2 := f.FilterSuggestions([]string{"дом", "xyzzyqqq"}, true)
		require.True(t, ok2)
		require.Contains(t, kept2, "дом")
		require.NotContains(t, kept2, "xyzzyqqq")
	}
}

func TestRussianSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ru.RussianSuppressMisspelledSuggestionsFilter"))
}
