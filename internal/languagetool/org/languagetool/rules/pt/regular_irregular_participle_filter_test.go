package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestIsRegularParticiple(t *testing.T) {
	require.True(t, IsRegularParticiple("assado"))
	require.True(t, IsRegularParticiple("assadas"))
	require.False(t, IsRegularParticiple("aceite"))
}

func TestRegularIrregularParticipleFilter_Suggest(t *testing.T) {
	f := NewRegularIrregularParticipleFilter()
	got := f.Suggest("RegularToIrregular", "aceitado", []string{"aceite", "aceitado"}, "{suggestion}")
	require.Equal(t, "aceite", got)
	got = f.Suggest("IrregularToRegular", "aceite", []string{"aceite", "aceitado"}, "{suggestion}")
	require.Equal(t, "aceitado", got)
}

func TestRegularIrregularParticipleFilter_AcceptRuleMatch(t *testing.T) {
	f := NewRegularIrregularParticipleFilter()
	// without synth → nil
	pos := "VMP00SM"
	lem := "gastar"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("gastado", &pos, &lem), 0)
	m := rules.NewRuleMatch(nil, nil, 0, 7, "msg")
	m.SetSuggestedReplacements([]string{"{suggestion}"})
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil))

	f.Synthesize = func(lemma, desired string) []string {
		return []string{"gasto", "gastado"}
	}
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"gasto"}, out.GetSuggestedReplacements())
}

func TestPTParticipleAndPartialRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.RegularIrregularParticipleFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.NoDisambiguationPortuguesePartialPosTagFilter"))
}
