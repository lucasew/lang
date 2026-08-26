package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFRSuppressMisspelledRegistered(t *testing.T) {
	class := "org.languagetool.rules.fr.FrenchSuppressMisspelledSuggestionsFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class))
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	// without speller: keep all suggestions (Java null SpellingCheckRule)
	m := rules.NewRuleMatch(rules.NewFakeRule("X"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"bon", "xyzzy"})
	out := f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"bon", "xyzzy"}, out.GetSuggestedReplacements())
}

func TestFRFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.fr.MakeContractionsFilter",
		"org.languagetool.rules.fr.DateCheckFilter",
		"org.languagetool.rules.fr.NewYearDateFilter",
		"org.languagetool.rules.fr.DMYDateCheckFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	}
}

func TestFRMakeContractionsFilter(t *testing.T) {
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.fr.MakeContractionsFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("C"), nil, 0, 5, "msg")
	m.SetSuggestedReplacements([]string{"de le livre", "à les amis"})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"du livre", "aux amis"}, out.GetSuggestedReplacements())
}
