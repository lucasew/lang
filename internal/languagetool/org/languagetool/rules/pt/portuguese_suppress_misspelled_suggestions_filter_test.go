package pt

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestPortugueseSuppressMisspelledSuggestionsFilter_NoDictKeepsAll(t *testing.T) {
	ClearPortugueseFilterSpeller()
	f := NewPortugueseSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"casa", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"casa", "xyz"}, kept)
}

func TestPortugueseSuppressMisspelledSuggestionsFilter_Injected(t *testing.T) {
	ClearPortugueseFilterSpeller()
	f := NewPortugueseSuppressMisspelledSuggestionsFilter()
	f.IsMisspelled = func(w string) bool { return w == "xyz" }
	kept, ok := f.FilterSuggestions([]string{"casa", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"casa"}, kept)
	_, ok = f.FilterSuggestions([]string{"xyz"}, true)
	require.False(t, ok)
}

func TestPortugueseSuppressMisspelledSuggestionsFilter_AcceptRuleMatch(t *testing.T) {
	ClearPortugueseFilterSpeller()
	f := NewPortugueseSuppressMisspelledSuggestionsFilter()
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

func TestPortugueseSuppressMisspelledSuggestionsFilter_WithPTDict(t *testing.T) {
	ClearPortugueseFilterSpeller()
	t.Cleanup(ClearPortugueseFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java getDictFilename: pt-BR default, pt-PT-90, pt-PT-45
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/pt/src/main/resources/org/languagetool/resource/pt/spelling/pt-BR.dict"),
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/pt/src/main/resources/org/languagetool/resource/pt/spelling/pt-PT-90.dict"),
		filepath.Join(root, "third_party/pt/spelling/pt-BR.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WirePortugueseFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("pt spelling .dict not in tree (Java MorfologikPortugueseSpellerRule)")
	}
	f := NewPortugueseSuppressMisspelledSuggestionsFilter()
	kept, ok := f.FilterSuggestions([]string{"casa", "xyzzyqqq"}, true)
	require.True(t, ok)
	// junk must be misspelled
	require.NotContains(t, kept, "xyzzyqqq")
}

func TestPortugueseSuppressMisspelledSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.PortugueseSuppressMisspelledSuggestionsFilter"))
}
