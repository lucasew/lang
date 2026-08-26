package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestAdvancedSynthesizerFilter_FailClosedWithoutSynth(t *testing.T) {
	ClearDefaultSynthesize()
	// Ensure no process-wide English synthesizer from other tests.
	patterns.RegisterLanguageSynthesizer("en", nil)
	// RegisterLanguageSynthesizer ignores nil — clear via overwrite only if API allows.
	// Use a fresh filter with no wire; fail-closed must not invent forms.
	f := NewAdvancedSynthesizerFilter()
	// If some other package registered en synth, SetSynthesize(nil) forces fail-closed.
	if f.Synthesize != nil {
		f.SetSynthesize(nil)
	}
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

func TestAdvancedSynthesizerFilter_EmptySynthFormsReturnsMatch(t *testing.T) {
	// Java: replacements.length == 0 → return match (no invent forms).
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
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
	m.SetSuggestedReplacements([]string{"orig"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "VBG",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.Equal(t, m, out)
	require.Equal(t, []string{"orig"}, out.GetSuggestedReplacements())
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

func TestAdvancedSynthesizerFilter_LanguageRegistrySynth(t *testing.T) {
	// Java language.getSynthesizer() path via patterns.LanguageSynthesizer("en").
	ClearDefaultSynthesize()
	prev := patterns.LanguageSynthesizer("en")
	t.Cleanup(func() {
		ClearDefaultSynthesize()
		if prev != nil {
			patterns.RegisterLanguageSynthesizer("en", prev)
		} else {
			// Cannot fully unregister; leave empty synth (no invent forms).
			patterns.RegisterLanguageSynthesizer("en", synthesis.FuncSynthesizer{})
		}
	})
	patterns.RegisterLanguageSynthesizer("en", synthesis.FuncSynthesizer{
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
			require.True(t, posTagRegExp)
			if token != nil && token.GetLemma() != nil && *token.GetLemma() == "be" && posTag == "VBD" {
				return []string{"was", "were"}, nil
			}
			return nil, nil
		},
	})
	f := NewAdvancedSynthesizerFilter()
	// Force registry path (no inject).
	f.SetSynthesize(nil)
	f.ensureSynth()
	require.NotNil(t, f.Synthesize, "ensureSynth must attach LanguageSynthesizer path")
	lemma := "be"
	pos1 := "VB"
	pos2 := "VBD"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("be", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("ran", &pos2, nil))
	t2.SetStartPos(3)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "VBD",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"was", "were"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_MissingRequiredArgPanics(t *testing.T) {
	// Java getRequired throws IllegalArgumentException when key absent.
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string { return []string{"x"} })
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x",
			// postagSelect missing
		}, 0, nil, nil)
	})
	// Missing required still panics even without synthesizer (Java order).
	f2 := NewAdvancedSynthesizerFilter()
	f2.SetSynthesize(nil)
	require.Panics(t, func() {
		_ = f2.AcceptRuleMatch(m, map[string]string{
			"lemmaFrom": "1", "postagFrom": "1", "postagSelect": "N",
			// lemmaSelect missing
		}, 0, nil, nil)
	})
}

func TestAdvancedSynthesizerFilter_NewLemmaOverridesNullLemma(t *testing.T) {
	// Java: newLemma non-underscore replaces null desiredLemma before null check.
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		if lemma == "fixed" {
			return []string{"fixed-form"}
		}
		return nil
	})
	pos1 := "VB"
	// lemma pointer nil on reading
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", &pos1, nil))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "VB", "postagSelect": "VB",
		"newLemma": "fixed",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fixed-form"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.AdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.AdvancedSynthesizerFilter"))
}
