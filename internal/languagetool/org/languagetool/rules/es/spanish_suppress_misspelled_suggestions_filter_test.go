package es

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestSpanishSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewSpanishSuppressMisspelledSuggestionsFilter()
	// no dict: FilterDictIsMisspelled returns false (Java null SpellingCheckRule)
	kept, ok := f.FilterSuggestions([]string{"casa", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"casa", "xyz"}, kept)
}

func TestSpanishSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewSpanishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"casa", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"casa"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestSpanishSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewSpanishSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"casa", "xyz"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa"}, out.GetSuggestedReplacements())

	m2 := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 4, "msg")
	m2.SetSuggestedReplacements([]string{"xyz"})
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"suppressMatch": "true"}, 0, nil, nil))
}

func TestSpanishSuppressMisspelledSuggestionsFilter_WithESDict(t *testing.T) {
	ClearSpanishFilterSpeller()
	t.Cleanup(ClearSpanishFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/resource/es/es-ES.dict"),
		filepath.Join(root, "third_party/es/es-ES.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireSpanishFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("es-ES.dict not in tree (Java Spanish default spelling)")
	}
	f := NewSpanishSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"casa", "xyzzyqqq"}, true)
	require.True(t, ok)
	require.Contains(t, kept, "casa")
	require.NotContains(t, kept, "xyzzyqqq")
}

func TestSpanishSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.es.SpanishSuppressMisspelledSuggestionsFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.es.SpanishSuppressMisspelledSuggestionsFilter"))
}
