package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDecadeSpellingFilter(t *testing.T) {
	f := NewDecadeSpellingFilter()
	// 1990 → decade 90, century 19 → wiek XX
	msg := f.FormatMessage("lata {dekada}. wieku {wiek}", "1990")
	require.Equal(t, "lata 90. wieku XX", msg)
	// 2000 → century 20 → XXI
	msg = f.FormatMessage("{dekada}/{wiek}", "2000")
	require.Equal(t, "00/XXI", msg)
	// century needs 2 digits; single char unparseable
	require.Equal(t, "", f.FormatMessage("x", "1"))
	// non-numeric century
	require.Equal(t, "", f.FormatMessage("x", "ab90"))
}

func TestDecadeSpellingFilter_AcceptRuleMatch(t *testing.T) {
	f := NewDecadeSpellingFilter()
	m := rules.NewRuleMatch(nil, nil, 3, 10, "lata {dekada}. wieku {wiek}")
	m.ShortMessage = "dekada"
	out := f.AcceptRuleMatch(m, map[string]string{"lata": "1990"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "lata 90. wieku XX", out.GetMessage())
	require.Equal(t, 3, out.GetFromPos())
	require.Equal(t, 10, out.GetToPos())
	require.Equal(t, "dekada", out.ShortMessage)

	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"lata": "xx90"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(nil, map[string]string{"lata": "1990"}, 0, nil, nil))
}

func TestPLDecadeSpellingFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.pl.DecadeSpellingFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	m := rules.NewRuleMatch(nil, nil, 0, 1, "{dekada}/{wiek}")
	out := f.AcceptRuleMatch(m, map[string]string{"lata": "2000"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "00/XXI", out.GetMessage())
}
