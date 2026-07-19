package de

// Twin of GermanNumberInWordFilterTest — speller-gated digit-in-word suggestions (Java).
import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestGermanNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewGermanNumberInWordFilter()
	// Without speller dict: no invent candidates
	require.Empty(t, f.Suggestions("H0use"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "H0use"}, 0, nil, nil))
}

func TestGermanNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	f := NewGermanNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool {
		return w != "House" && w != "Hus"
	}
	f.inner.GetSuggestions = nil
	// Gate logic on AbstractNumberInWordFilter (inner); public Suggestions requires dict.
	got := f.inner.Suggestions("H0use")
	require.Contains(t, got, "House")
	got2 := f.inner.Suggestions("H4us")
	require.Contains(t, got2, "Hus")
	require.Empty(t, f.inner.Suggestions("Haus"))
}

func TestGermanNumberInWordFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewGermanNumberInWordFilter()
	// Java GermanNumberInWordFilterTest: Aut0r → Autor
	sugg := f.Suggestions("Aut0r")
	require.Contains(t, sugg, "Autor")
	// Grüß0e → Grüße (0→o)
	sugg2 := f.Suggestions("Grüß0e")
	require.Contains(t, sugg2, "Grüße")
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "fake msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "Aut0r"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "Autor")
}

func TestGermanNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.de.GermanNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
