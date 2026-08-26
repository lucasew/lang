package nl

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDutchNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearDutchFilterSpeller()
	f := NewDutchNumberInWordFilter()
	require.Empty(t, f.Suggestions("w0ord"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "w0ord"}, 0, nil, nil))
}

func TestDutchNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	f := NewDutchNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool { return w != "woord" && w != "wrd" }
	f.inner.GetSuggestions = nil
	// Gate logic on AbstractNumberInWordFilter (inner); public Suggestions requires dict.
	// w0ord → 0→o = woord (known); strip digits = wrd (known) — both when not misspelled
	got := f.inner.Suggestions("w0ord")
	require.Contains(t, got, "woord")
}

func TestDutchNumberInWordFilter_WithNLDict(t *testing.T) {
	ClearDutchFilterSpeller()
	t.Cleanup(ClearDutchFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	// Java MorfologikDutchSpellerRule: /nl/spelling/nl_NL.dict
	candidates := []string{
		filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/nl/src/main/resources/org/languagetool/resource/nl/spelling/nl_NL.dict"),
		filepath.Join(root, "third_party/nl/spelling/nl_NL.dict"),
		filepath.Join(root, "third_party/dutch-pos-dict/org/languagetool/resource/nl/spelling/nl_NL.dict"),
	}
	wired := false
	for _, dict := range candidates {
		if WireDutchFilterSpeller(dict) {
			wired = true
			break
		}
	}
	if !wired {
		t.Skip("nl_NL.dict not in tree (Java MorfologikDutchSpellerRule /nl/spelling/nl_NL.dict)")
	}
	f := NewDutchNumberInWordFilter()
	sugg := f.Suggestions("w0ord")
	require.NotEmpty(t, sugg)
	require.Contains(t, sugg, "woord")
}

func TestDutchNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.nl.DutchNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
