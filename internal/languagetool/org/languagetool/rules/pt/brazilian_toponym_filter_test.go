package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestBrazilianToponymMap(t *testing.T) {
	m := LoadBrazilianToponymMap()
	require.True(t, m.IsValidToponym("São Paulo"))
	require.True(t, m.IsValidToponym("Venho do Rio de Janeiro")) // suffix match
	require.False(t, m.IsValidToponym("Narnia"))
	require.True(t, m.IsToponymInState("são paulo", "SP"))
}

func TestBrazilianToponymFilter(t *testing.T) {
	f := NewBrazilianToponymFilter()
	require.Equal(t, "–SP", f.Suggest("São Paulo", "-", "SP"))
	require.Equal(t, "", f.Suggest("São Paulo", "–SP", "SP"))
	require.Equal(t, "", f.Suggest("Narnia", "-", "XX"))
}

func TestBrazilianToponymFilter_AcceptRuleMatch(t *testing.T) {
	f := NewBrazilianToponymFilter()
	m := rules.NewRuleMatch(nil, nil, 0, 5, "msg")
	// groups: full, toponym, underlined, state
	out := f.AcceptRuleMatch(m, nil, nil, []string{
		"Niterói (RJ)", "Niterói", " (RJ)", "RJ",
	})
	require.NotNil(t, out)
	require.Equal(t, []string{"–RJ"}, out.GetSuggestedReplacements())

	// already correct en-dash form → drop
	require.Nil(t, f.AcceptRuleMatch(m, nil, nil, []string{
		"Niterói–RJ", "Niterói", "–RJ", "RJ",
	}))
	// invalid toponym → drop
	require.Nil(t, f.AcceptRuleMatch(m, nil, nil, []string{
		"Oogabooga/RJ", "Oogabooga", "/RJ", "RJ",
	}))
}

func TestBrazilianToponymFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRegexRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.BrazilianToponymFilter"))
}
