package filters

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestArabicAdvancedSynthesizerFilter_FailClosedWithoutSynth(t *testing.T) {
	ClearDefaultSynthesize()
	// Use a fresh filter with no wire; fail-closed must not invent forms.
	f := NewArabicAdvancedSynthesizerFilter()
	// If some other package registered ar synth, SetSynthesize(nil) forces fail-closed.
	if f.Synthesize != nil {
		f.SetSynthesize(nil)
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x", "postagSelect": "N",
	}, 0, nil, nil))
}

func TestArabicAdvancedSynthesizerFilter_WithInjectedSynth(t *testing.T) {
	ClearDefaultSynthesize()
	f := NewArabicAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		if lemma == "كتب" && postag == "V" {
			return []string{"يكتب"}
		}
		return nil
	})
	lemma := "كتب"
	pos1 := "N"
	pos2 := "V"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("كتاب", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("يقرأ", &pos2, nil))
	t2.SetStartPos(5)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		// Java lemmaSelect is a POS regex (not lemma surface).
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "N", "postagSelect": "V",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"يكتب"}, out.GetSuggestedReplacements())
}

func TestArabicAdvancedSynthesizerFilter_EmptySynthFormsReturnsMatch(t *testing.T) {
	// Java: replacements.length == 0 → return match (no invent forms).
	ClearDefaultSynthesize()
	f := NewArabicAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		return nil
	})
	lemma := "كتب"
	pos1 := "N"
	pos2 := "V"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("كتاب", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("يقرأ", &pos2, nil))
	t2.SetStartPos(5)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	m.SetSuggestedReplacements([]string{"orig"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "N", "postagSelect": "V",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.Equal(t, m, out)
	require.Equal(t, []string{"orig"}, out.GetSuggestedReplacements())
}

func TestArabicAdvancedSynthesizerFilter_ProcessWideWire(t *testing.T) {
	ClearDefaultSynthesize()
	t.Cleanup(ClearDefaultSynthesize)
	WireDefaultSynthesize(func(lemma, postag string) []string {
		if lemma == "مشي" {
			return []string{"يمشي"}
		}
		return nil
	})
	f := NewArabicAdvancedSynthesizerFilter()
	lemma := "مشي"
	pos1 := "V"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("مشي", &pos1, &lemma))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "V", "postagSelect": "V",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"يمشي"}, out.GetSuggestedReplacements())
}

func TestArabicAdvancedSynthesizerFilter_LanguageRegistrySynth(t *testing.T) {
	// Java language.getSynthesizer() path via patterns.LanguageSynthesizer("ar").
	ClearDefaultSynthesize()
	prev := patterns.LanguageSynthesizer("ar")
	t.Cleanup(func() {
		ClearDefaultSynthesize()
		if prev != nil {
			patterns.RegisterLanguageSynthesizer("ar", prev)
		} else {
			// Cannot fully unregister; leave empty synth (no invent forms).
			patterns.RegisterLanguageSynthesizer("ar", synthesis.FuncSynthesizer{})
		}
	})
	patterns.RegisterLanguageSynthesizer("ar", synthesis.FuncSynthesizer{
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
			require.True(t, posTagRegExp)
			if token != nil && token.GetLemma() != nil && *token.GetLemma() == "كان" && posTag == "V" {
				return []string{"يكون", "كانت"}, nil
			}
			return nil, nil
		},
	})
	f := NewArabicAdvancedSynthesizerFilter()
	// Force registry path (no inject).
	f.SetSynthesize(nil)
	f.ensureSynth()
	require.NotNil(t, f.Synthesize, "ensureSynth must attach LanguageSynthesizer path")
	lemma := "كان"
	pos1 := "N"
	pos2 := "V"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("كان", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("فعل", &pos2, nil))
	t2.SetStartPos(4)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "N", "postagSelect": "V",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"يكون", "كانت"}, out.GetSuggestedReplacements())
}

func TestArabicAdvancedSynthesizerFilter_MissingRequiredArgPanics(t *testing.T) {
	// Java getRequired throws IllegalArgumentException when key absent.
	ClearDefaultSynthesize()
	f := NewArabicAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string { return []string{"x"} })
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x",
			// postagSelect missing
		}, 0, nil, nil)
	})
	// Missing required still panics even without synthesizer (Java order).
	f2 := NewArabicAdvancedSynthesizerFilter()
	f2.SetSynthesize(nil)
	require.Panics(t, func() {
		_ = f2.AcceptRuleMatch(m, map[string]string{
			"lemmaFrom": "1", "postagFrom": "1", "postagSelect": "N",
			// lemmaSelect missing
		}, 0, nil, nil)
	})
}

func TestArabicAdvancedSynthesizerFilter_NewLemmaOverridesNullLemma(t *testing.T) {
	// Java: newLemma non-underscore replaces null desiredLemma before null check.
	ClearDefaultSynthesize()
	f := NewArabicAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		if lemma == "fixed" {
			return []string{"fixed-form"}
		}
		return nil
	})
	pos1 := "V"
	// lemma pointer nil on reading
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", &pos1, nil))
	t1.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 1, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "V", "postagSelect": "V",
		"newLemma": "fixed",
	}, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fixed-form"}, out.GetSuggestedReplacements())
}

func TestArabicAdvancedSynthesizerFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter"))
}

func TestAdvancedSynthFilterConstruct(t *testing.T) {
	// Auxiliary construct smoke (outcomes covered above).
	ClearDefaultSynthesize()
	f := NewArabicAdvancedSynthesizerFilter()
	f.SetSynthesize(func(lemma, postag string) []string {
		return []string{lemma + ":" + postag}
	})
	require.NotNil(t, f.AbstractAdvancedSynthesizerFilter)
	require.NotNil(t, f.Synthesize)
}
