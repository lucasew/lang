package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestAdvancedSynthesizerFilter_FailClosedWithoutSynth(t *testing.T) {
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x", "postagSelect": "N",
	}, 0, nil, nil))
}

func TestAdvancedSynthesizerFilter_WithInjectedSynth(t *testing.T) {
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		// desiredPostag is the POS from the selected reading (not the select regex)
		if lemma == "casa" && postag == "NCFS000" {
			return []string{"casa"}
		}
		return nil
	})
	lemma := "casa"
	pos1 := "NCFS000"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &pos1, &lemma))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "casa", "postagSelect": "NCFS000",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casa"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.AdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.ca.AdvancedSynthesizerFilter"))
}
