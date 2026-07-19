package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestConfusionCheckFilter_Embedded(t *testing.T) {
	f := NewConfusionCheckFilter()
	require.NotEmpty(t, f.Pairs)
	res := f.Suggest("acaro", "NCMS000", "", "se escribe con tilde", "{suggestion}")
	require.True(t, res.OK)
	require.Equal(t, "ácaro", res.Replacement)
}

func TestConfusionCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewConfusionCheckFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("C"), nil, 0, 5, "se escribe con tilde")
	m.SetSuggestedReplacements([]string{"{suggestion}"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"postag": "NC.*",
		"form":   "acaro",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"ácaro"}, out.GetSuggestedReplacements())
}
