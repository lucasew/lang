package ca

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestCatalanNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewCatalanNumberInWordFilter()
	require.Empty(t, f.Suggestions("cas4"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "cas4"}, 0, nil, nil))
}

func TestCatalanNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	f := NewCatalanNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool { return w != "cas" }
	f.inner.GetSuggestions = nil
	// Gate logic on AbstractNumberInWordFilter (inner); public Suggestions requires dict.
	require.Equal(t, []string{"cas"}, f.inner.Suggestions("cas4"))
}

func TestCatalanNumberInWordFilter_WithCADict(t *testing.T) {
	ClearCatalanFilterSpeller()
	t.Cleanup(ClearCatalanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java MorfologikCatalanSpellerRule: /ca/ca-ES_spelling.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/ca/src/main/resources/org/languagetool/resource/ca/ca-ES_spelling.dict"),
		filepath.Join(root, "third_party/ca/ca-ES_spelling.dict"),
		filepath.Join(root, "third_party/catalan-pos-dict/org/languagetool/resource/ca/ca-ES_spelling.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireCatalanFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("ca-ES_spelling.dict not in tree (Java MorfologikCatalanSpellerRule)")
	}
	f := NewCatalanNumberInWordFilter()
	sugg := f.Suggestions("cas4")
	// strip digits → cas when known
	require.NotEmpty(t, sugg)
}

func TestCatalanNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.ca.CatalanNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}

func TestCatalanSuppressMisspelled(t *testing.T) {
	ClearCatalanFilterSpeller()
	f := NewCatalanSuppressMisspelledSuggestionsFilter()
	// Without dict, Java null speller → all misspelled; inject override for gate test.
	f.SetIsMisspelled(func(w string) bool { return w == "xyz" })
	kept, ok := f.FilterSuggestions([]string{"bé", "xyz"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"bé"}, kept)
}
