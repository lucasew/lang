package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestPotentialCompoundFilter_FailClosedWithoutSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	// Without speller hook: fail-closed → hyphenated only (do not invent joined valid).
	s := f.Suggestions("Haus", "tür")
	require.Equal(t, []string{"Haus-Tür"}, s)
}

func TestPotentialCompoundFilter_WithInjectedSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewPotentialCompoundFilter()
	// Joined valid via speller twin
	f.JoinedIsValid = func(string) bool { return true }
	s2 := f.Suggestions("Haus", "tür")
	require.Contains(t, s2, "Haustür")
	f.JoinedIsValid = func(string) bool { return false }
	s3 := f.Suggestions("Haus", "Tür")
	require.Equal(t, []string{"Haus-Tür"}, s3)
}

func TestPotentialCompoundFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewPotentialCompoundFilter()
	// No override: uses FilterDict like Java default spelling rule.
	// Unknown junk → hyphenated only
	s := f.Suggestions("Xyzzy", "qqqq")
	require.Equal(t, []string{"Xyzzy-Qqqq"}, s)
	// Abendessen is typically in de_DE as a compound form — if known, suggest joined
	s2 := f.Suggestions("Abend", "essen")
	if FilterDictIsMisspelled("Abendessen") {
		require.Equal(t, []string{"Abend-Essen"}, s2)
	} else {
		require.Contains(t, s2, "Abendessen")
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"part1": "Xyzzy", "part2": "qqqq"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Xyzzy-Qqqq"}, out.GetSuggestedReplacements())
}

func TestPotentialCompoundFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.PotentialCompoundFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.PotentialCompoundFilter"))
}
