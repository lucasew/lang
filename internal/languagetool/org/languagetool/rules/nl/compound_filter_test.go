package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Port of CompoundFilterTest.testFilter
func TestCompoundFilter_Filter(t *testing.T) {
	f := NewCompoundFilter()
	cases := []struct {
		words []string
		want  string
	}{
		{[]string{"tv", "meubel"}, "tv-meubel"},
		{[]string{"test-tv", "meubel"}, "test-tv-meubel"},
		{[]string{"onzin", "tv"}, "onzin-tv"},
		{[]string{"auto", "onderdeel"}, "auto-onderdeel"},
		{[]string{"test", "e-mail"}, "test-e-mail"},
		{[]string{"taxi", "jongen"}, "taxi-jongen"},
		{[]string{"rij", "instructeur"}, "rijinstructeur"},
		{[]string{"ANWB", "wagen"}, "ANWB-wagen"},
		{[]string{"pro-deo", "advocaat"}, "pro-deoadvocaat"},
		{[]string{"ANWB", "tv", "wagen"}, "ANWB-tv-wagen"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, f.Suggest(tc.words), "words=%v", tc.words)
	}
}

func TestCompoundFilter_AcceptRuleMatchRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.nl.CompoundFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.nl.CompoundFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("C"), nil, 0, 10, "use <suggestion>x</suggestion>")
	m.ShortMessage = "short <suggestion>y</suggestion>"
	out := f.AcceptRuleMatch(m, map[string]string{"word1": "tv", "word2": "meubel"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"tv-meubel"}, out.GetSuggestedReplacements())
	require.Contains(t, out.GetMessage(), "tv-meubel")
	require.Contains(t, out.ShortMessage, "tv-meubel")
}
