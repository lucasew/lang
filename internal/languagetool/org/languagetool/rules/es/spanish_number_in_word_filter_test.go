package es

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestSpanishNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearSpanishFilterSpeller()
	f := NewSpanishNumberInWordFilter()
	// without speller: fail-closed (no invent)
	require.Empty(t, f.Suggestions("cas4"))
	require.Empty(t, f.Suggestions("t0d0"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "cas4"}, 0, nil, nil))
}

func TestSpanishNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	f := NewSpanishNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool {
		return w != "cas" && w != "todo" && w != "td"
	}
	f.inner.GetSuggestions = nil
	// Gate logic on AbstractNumberInWordFilter (inner); public Suggestions requires dict.
	require.Nil(t, f.inner.Suggestions("hola"))
	require.Equal(t, []string{"cas"}, f.inner.Suggestions("cas4"))
	require.Equal(t, []string{"todo", "td"}, f.inner.Suggestions("t0d0"))
	// all-digit word: only 0→o form when known
	f.inner.IsMisspelled = func(w string) bool { return w != "o23" }
	require.Equal(t, []string{"o23"}, f.inner.Suggestions("023"))
}

func TestSpanishNumberInWordFilter_AcceptRuleMatch_Injected(t *testing.T) {
	f := NewSpanishNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool { return w != "cas" }
	f.inner.GetSuggestions = nil
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	// Public Accept fails closed without dict; exercise abstract path via inner.
	out := f.inner.AcceptRuleMatch(m, map[string]string{"word": "cas4"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"cas"}, out.GetSuggestedReplacements())
}

func TestSpanishNumberInWordFilter_WithESDict(t *testing.T) {
	ClearSpanishFilterSpeller()
	t.Cleanup(ClearSpanishFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Official path if vendored; skip when es.dict not in tree (not invent).
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/resource/es/es-ES.dict"),
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/es/src/main/resources/org/languagetool/resource/es/hunspell/es.dict"),
		filepath.Join(root, "third_party/es/es-ES.dict"),
		filepath.Join(root, "third_party/es/hunspell/es.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireSpanishFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("es.dict not in tree (Java MorfologikSpanishSpellerRule resource)")
	}
	f := NewSpanishNumberInWordFilter()
	// 0→o when form is known in dict
	sugg := f.Suggestions("t0do")
	require.NotEmpty(t, sugg)
	require.Contains(t, sugg, "todo")
}

func TestESSpanishNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.es.SpanishNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
