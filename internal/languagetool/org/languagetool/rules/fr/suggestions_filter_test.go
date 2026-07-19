package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestSuggestionsFilter(t *testing.T) {
	f := NewSuggestionsFilter()
	got := f.Filter([]string{"bonjour", "xyz123", "salut"}, `.*\d.*`)
	require.Equal(t, []string{"bonjour", "salut"}, got)
	// Java Matcher.matches: substring-only pattern without anchors still full-matches via wrap
	got = f.Filter([]string{"ab", "abc"}, `ab`)
	require.Equal(t, []string{"abc"}, got) // "abc" does not fully match ^ab$
}

func TestSuggestionsFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.fr.SuggestionsFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.fr.SuggestionsFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 3, "msg")
	m.SetSuggestedReplacements([]string{"ok", "bad1"})
	out := f.AcceptRuleMatch(m, map[string]string{"RemoveSuggestionsRegexp": `.*\d.*`}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"ok"}, out.GetSuggestedReplacements())
}
