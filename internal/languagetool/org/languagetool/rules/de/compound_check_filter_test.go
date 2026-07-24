package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestCompoundCheckFilter(t *testing.T) {
	f := NewCompoundCheckFilter()
	require.True(t, f.Accept("Zeit", "Punkt"))
	require.True(t, f.Accept("zeit", "punkt"))
	require.False(t, f.Accept("xyz", "abc"))
}

func TestCompoundCheckFilter_AcceptRuleMatchRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.de.CompoundCheckFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.de.CompoundCheckFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("C"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"part1": "Zeit", "part2": "Punkt"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"part1": "no", "part2": "way"}, 0, nil, nil))
}
