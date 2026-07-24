package fr

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFrenchNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchNumberInWordFilter()
	require.Empty(t, f.Suggestions("m0t")) // fail-closed without speller
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "m0t"}, 0, nil, nil))
}

func TestFrenchNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	f := NewFrenchNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool { return w != "mot" && w != "mt" }
	f.inner.GetSuggestions = nil
	// Gate logic on AbstractNumberInWordFilter (inner); public Suggestions requires dict.
	require.Equal(t, []string{"mot", "mt"}, f.inner.Suggestions("m0t"))
}

func TestFrenchNumberInWordFilter_WithFRDict(t *testing.T) {
	ClearFrenchFilterSpeller()
	t.Cleanup(ClearFrenchFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java MorfologikFrenchSpellerRule: /fr/french.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/fr/src/main/resources/org/languagetool/resource/fr/french.dict"),
		filepath.Join(root, "third_party/fr/french.dict"),
		filepath.Join(root, "third_party/french-pos-dict/org/languagetool/resource/fr/french.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireFrenchFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("french.dict not in tree (Java MorfologikFrenchSpellerRule /fr/french.dict)")
	}
	f := NewFrenchNumberInWordFilter()
	// 0→o when form is known
	sugg := f.Suggestions("m0t")
	require.NotEmpty(t, sugg)
	require.Contains(t, sugg, "mot")
}

func TestFrenchNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.fr.FrenchNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}

func TestFrenchSuppressMisspelled(t *testing.T) {
	ClearFrenchFilterSpeller()
	f := NewFrenchSuppressMisspelledSuggestionsFilter()
	// no dict / FilterDictIsMisspelled false → keep all (Java null speller)
	kept, ok := f.FilterSuggestions([]string{"bon"}, true)
	require.True(t, ok)
	require.Equal(t, []string{"bon"}, kept)
}
