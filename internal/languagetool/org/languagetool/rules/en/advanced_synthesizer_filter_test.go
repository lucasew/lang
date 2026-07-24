package en

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
		if lemma == "go" && postag == "VBG" {
			return []string{"going"}
		}
		return nil
	})
	lemma := "go"
	pos1 := "VB"
	pos2 := "VBG"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("go", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("running", &pos2, nil))
	t2.SetStartPos(3)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		// Java lemmaSelect is a POS regex (not lemma surface).
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "VBG",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"going"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_ProcessWideWire(t *testing.T) {
	ClearDefaultSynthesize()
	t.Cleanup(ClearDefaultSynthesize)
	WireDefaultSynthesize(func(lemma, postag string) []string {
		if lemma == "walk" {
			return []string{"walked"}
		}
		return nil
	})
	f := NewAdvancedSynthesizerFilter()
	lemma := "walk"
	pos1 := "VBD"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("walk", &pos1, &lemma))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "VBD", "postagSelect": "VBD",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"walked"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.AdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.AdvancedSynthesizerFilter"))
}
