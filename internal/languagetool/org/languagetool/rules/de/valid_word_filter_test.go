package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestValidWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewValidWordFilter()
	// default without dict: always misspelled → keep match
	require.True(t, f.Accept("vielleicht", "der"))
	m := rules.NewRuleMatch(rules.NewFakeRule("V"), nil, 0, 4, "msg")
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"word1": "vielleicht", "word2": "der"}, 0, nil, nil))
}

func TestValidWordFilter_WithInjectedSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewValidWordFilter()
	f.IsMisspelled = func(w string) bool {
		return w != "Promotionsstudierende"
	}
	// Java example: "(Promotions)Studierende" → Promotionsstudierende is ok → suppress
	require.False(t, f.Accept("Promotions", "Studierende"))
	require.True(t, f.Accept("foo", "Bar"))
}

func TestValidWordFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewValidWordFilter()
	// No override: FilterDict like Java default spelling rule.
	// "vielleicht" alone is known; "vielleichtder" is not → keep match
	require.True(t, f.Accept("vielleicht", "der"))
	// Unknown junk stays as match
	require.True(t, f.Accept("xyzzy", "qqq"))
	// If a known compound is in the dict, suppress (Java: !isMisspelled → null)
	if !FilterDictIsMisspelled("Abendessen") {
		require.False(t, f.Accept("Abend", "essen"))
	}
}

func TestValidWordFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.ValidWordFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.ValidWordFilter"))
}
