package de

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRemoveUnknownCompoundsFilter_FailClosedWithoutSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewRemoveUnknownCompoundsFilter()
	// soft invent removed: without IsMisspelled/dict do not keep match
	require.False(t, f.Accept("Haus", "Tür"))
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"part1": "Haus", "part2": "Tür"}, 0, nil, nil))
}

func TestRemoveUnknownCompoundsFilter_WithSpeller(t *testing.T) {
	ClearGermanFilterSpeller()
	f := NewRemoveUnknownCompoundsFilter()
	f.IsMisspelled = func(w string) bool { return w == "Hausxyz" }
	require.True(t, f.Accept("Haus", "Tür"))  // Haus + tür not in misspelled set
	require.False(t, f.Accept("Haus", "Xyz")) // Hausxyz misspelled → drop
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"part1": "Haus", "part2": "Tür"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"part1": "Haus", "part2": "Xyz"}, 0, nil, nil))
}

func TestRemoveUnknownCompoundsFilter_WithDEDict(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	dict := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skipf("de_DE.dict not openable: %s", dict)
	}
	f := NewRemoveUnknownCompoundsFilter()
	// No override: uses FilterDict like Java GermanyGerman default spelling rule.
	// Haus is known; Haustür may or may not be in dict as one token — check known simple form.
	// part1+part2 lower: "Haus"+"tür" = "Haustür"
	// At least: unknown junk compound must drop
	require.False(t, f.Accept("Xyzzy", "Qqqq"))
	// "Haus"+"boot" style: if compound known keep — use injected for deterministic known form
	// Real dict: "vielleicht" is known; single-token compounds vary.
	// Java: drop when isMisspelled(compound). "xyzzyqq" style:
	require.False(t, f.Accept("xyz", "zyy"))
	// A form that is in the dict as a whole word should keep (probe Contains via Accept)
	// "Abend" + "essen" → "Abendessen" common compound — may be in de_DE.dict
	if f.Accept("Abend", "essen") {
		m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
		require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"part1": "Abend", "part2": "essen"}, 0, nil, nil))
	}
	// Explicit misspelled compound must drop
	require.Nil(t, f.AcceptRuleMatch(
		rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg"),
		map[string]string{"part1": "xyz", "part2": "zyy"}, 0, nil, nil))
}

func TestRemoveUnknownCompoundsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.RemoveUnknownCompoundsFilter"))
	// Direct type registered (wrapper optional)
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.de.RemoveUnknownCompoundsFilter")
	require.NotNil(t, f)
}
