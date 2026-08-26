package de

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
	// Force no synth even if german_synth.dict is discoverable (fail-closed path).
	f.SetSynthesize(nil)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x", "postagSelect": "N",
	}, 0, nil, nil))
}

func TestAdvancedSynthesizerFilter_WithInjectedSynth(t *testing.T) {
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		if lemma == "gehen" && postag == "VER:PA2:SIN:NEU" {
			return []string{"gegangen"}
		}
		return nil
	})
	lemma := "gehen"
	pos1 := "VER:INF:NON"
	pos2 := "VER:PA2:SIN:NEU"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("gehen", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("gemacht", &pos2, nil))
	t2.SetStartPos(6)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VER:INF:NON", "postagSelect": "VER:PA2:SIN:NEU",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"gegangen"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_ProcessWideWire(t *testing.T) {
	ClearDefaultSynthesize()
	t.Cleanup(ClearDefaultSynthesize)
	WireDefaultSynthesize(func(lemma, postag string) []string {
		if lemma == "haus" {
			return []string{"Haus"}
		}
		return nil
	})
	f := NewAdvancedSynthesizerFilter()
	lemma := "haus"
	pos1 := "SUB:NOM:SIN:NEU"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("haus", &pos1, &lemma))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "SUB:NOM:SIN:NEU", "postagSelect": "SUB:NOM:SIN:NEU",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Haus"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.AdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.AdvancedSynthesizerFilter"))
}
