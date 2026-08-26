package gl

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
	// Ensure no process-wide Galician synthesizer from other tests.
	patterns.RegisterLanguageSynthesizer("gl", nil)
	// RegisterLanguageSynthesizer ignores nil — clear via overwrite only if API allows.
	// Use a fresh filter with no wire; fail-closed must not invent forms.
	f := NewAdvancedSynthesizerFilter()
	// If some other package registered gl synth, SetSynthesize(nil) forces fail-closed.
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
		if lemma == "casa" && postag == "NC00000" {
			return []string{"casas"}
		}
		return nil
	})
	lemma := "casa"
	pos1 := "NC00000"
	pos2 := "NC00000"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casas", &pos2, nil))
	t2.SetStartPos(5)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		// Java lemmaSelect is a POS regex (not lemma surface).
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "NC00000", "postagSelect": "NC00000",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"casas"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_EmptySynthFormsReturnsMatch(t *testing.T) {
	// Java: replacements.length == 0 → return match (no invent forms).
	ClearDefaultSynthesize()
	f := NewAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		return nil
	})
	lemma := "casa"
	pos1 := "NC00000"
	pos2 := "NC00000"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casas", &pos2, nil))
	t2.SetStartPos(5)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	m.SetSuggestedReplacements([]string{"orig"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "NC00000", "postagSelect": "NC00000",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.Equal(t, m, out)
	require.Equal(t, []string{"orig"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_ProcessWideWire(t *testing.T) {
	ClearDefaultSynthesize()
	t.Cleanup(ClearDefaultSynthesize)
	WireDefaultSynthesize(func(lemma, postag string) []string {
		if lemma == "falar" {
			return []string{"falou"}
		}
		return nil
	})
	f := NewAdvancedSynthesizerFilter()
	lemma := "falar"
	pos1 := "VMIS3S0"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("falar", &pos1, &lemma))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "VMIS3S0", "postagSelect": "VMIS3S0",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"falou"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_LanguageRegistrySynth(t *testing.T) {
	// Java language.getSynthesizer() path via patterns.LanguageSynthesizer("gl").
	ClearDefaultSynthesize()
	prev := patterns.LanguageSynthesizer("gl")
	t.Cleanup(func() {
		ClearDefaultSynthesize()
		if prev != nil {
			patterns.RegisterLanguageSynthesizer("gl", prev)
		} else {
			// Cannot fully unregister; leave empty synth (no invent forms).
			patterns.RegisterLanguageSynthesizer("gl", synthesis.FuncSynthesizer{})
		}
	})
	patterns.RegisterLanguageSynthesizer("gl", synthesis.FuncSynthesizer{
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
			require.True(t, posTagRegExp)
			if token != nil && token.GetLemma() != nil && *token.GetLemma() == "ser" && posTag == "VMIS3S0" {
				return []string{"foi", "era"}, nil
			}
			return nil, nil
		},
	})
	f := NewAdvancedSynthesizerFilter()
	// Force registry path (no inject).
	f.SetSynthesize(nil)
	f.ensureSynth()
	require.NotNil(t, f.Synthesize, "ensureSynth must attach LanguageSynthesizer path")
	lemma := "ser"
	pos1 := "VMN0000"
	pos2 := "VMIS3S0"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("ser", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("foi", &pos2, nil))
	t2.SetStartPos(4)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VMN0000", "postagSelect": "VMIS3S0",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"foi", "era"}, out.GetSuggestedReplacements())
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
	pos1 := "NC00000"
	// lemma pointer nil on reading
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", &pos1, nil))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "NC00000", "postagSelect": "NC00000",
		"newLemma": "fixed",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fixed-form"}, out.GetSuggestedReplacements())
}

func TestAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.gl.AdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.gl.AdvancedSynthesizerFilter"))
}
