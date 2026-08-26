package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestSplitPartsByHyphen(t *testing.T) {
	require.Equal(t, []string{"Wacht", "ums", "pistole"},
		splitPartsByHyphen([]string{"Wacht", "ums-pistole"}))
	require.Equal(t, []string{"a", "b"}, splitPartsByHyphen([]string{"a", "b"}))
}

func TestRestoreRemovedHyphens(t *testing.T) {
	// Implementierungs + pflicht from "Implementierungs-pflicht"
	got := restoreRemovedHyphens([]string{"Implementierungs", "pflicht"}, "Implementierungs-pflicht")
	require.Equal(t, []string{"Implementierungs-", "pflicht"}, got)
	// multi-hyphen: Auto-Bahn-Netz
	got2 := restoreRemovedHyphens([]string{"Auto", "Bahn", "Netz"}, "Auto-Bahn-Netz")
	require.Equal(t, []string{"Auto-", "Bahn-", "Netz"}, got2)
}

func TestMatch_TypeUnknownWordAndShortMessage(t *testing.T) {
	ClearGermanFilterSpeller()
	t.Cleanup(ClearGermanFilterSpeller)
	dict := deHunspellPath(t, "de_DE.dict")
	if !WireGermanFilterSpeller(dict) {
		t.Skip("no dict")
	}
	r := NewGermanSpellerRule(map[string]string{
		"desc_spelling_short": "Tippfehler",
	})
	ms := r.Match(languagetool.AnalyzePlain("Das ist xyzzyqqq."))
	require.NotEmpty(t, ms)
	require.Equal(t, rules.RuleMatchTypeUnknownWord, ms[0].GetType())
	require.Equal(t, "Tippfehler", ms[0].GetShortMessage())
}
