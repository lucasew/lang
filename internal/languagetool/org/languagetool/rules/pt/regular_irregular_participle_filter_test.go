package pt

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	"github.com/stretchr/testify/require"
)

// Java isRegular matrix: lower ends with do|dos|da|das.
func TestIsRegularParticiple(t *testing.T) {
	// regular
	require.True(t, IsRegularParticiple("assado"))
	require.True(t, IsRegularParticiple("assadas"))
	require.True(t, IsRegularParticiple("gastado"))
	require.True(t, IsRegularParticiple("GASTADO"))
	require.True(t, IsRegularParticiple("gastados"))
	require.True(t, IsRegularParticiple("gastada"))
	require.True(t, IsRegularParticiple("gastadas"))
	// irregular (Java table negatives)
	require.False(t, IsRegularParticiple("aceite"))
	require.False(t, IsRegularParticiple("gasto"))
	require.False(t, IsRegularParticiple("impresso"))
	require.False(t, IsRegularParticiple("limpo"))
	require.False(t, IsRegularParticiple("pago"))
}

// controlledSynth builds a PortugueseSynthesizer twin backed by ManualSynthesizer
// so SynthesizeRE(token, postag, true) yields known dual participles (gastado/gasto style).
func controlledParticipleSynth(t *testing.T, lines string) *ptsynthesis.PortugueseSynthesizer {
	t.Helper()
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(lines))
	require.NoError(t, err)
	return ptsynthesis.NewPortugueseSynthesizer(manual)
}

func vmpToken(surface, lemma, postag string, startPos int) *languagetool.AnalyzedTokenReadings {
	pos := postag
	lem := lemma
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(surface, &pos, &lem), startPos)
}

func baseMatch(from, to int, msg, short, template string) *rules.RuleMatch {
	m := rules.NewRuleMatch(rules.NewFakeRule("PART"), nil, from, to, msg)
	m.SetShortMessage(short)
	m.SetType(rules.RuleMatchTypeHint)
	m.SetSuggestedReplacements([]string{template})
	return m
}

func TestRegularIrregularParticipleFilter_RegularToIrregular(t *testing.T) {
	// Java: RegularToIrregular + isRegular(token) → first non-regular among [0],[1]
	synth := controlledParticipleSynth(t,
		"gasto\tgastar\tVMP00SM\n"+
			"gastado\tgastar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)

	tok := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 7, "use irregular", "short", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"gasto"}, out.GetSuggestedReplacements())
	require.Equal(t, "use irregular", out.GetMessage())
	require.Equal(t, "short", out.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeHint, out.GetType())
}

func TestRegularIrregularParticipleFilter_IrregularToRegular(t *testing.T) {
	// Java: IrregularToRegular + !isRegular(token) → first regular among [0],[1]
	synth := controlledParticipleSynth(t,
		"gasto\tgastar\tVMP00SM\n"+
			"gastado\tgastar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)

	tok := vmpToken("gasto", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 5, "use regular", "s", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "IrregularToRegular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"gastado"}, out.GetSuggestedReplacements())
}

func TestRegularIrregularParticipleFilter_TemplatePlaceholders(t *testing.T) {
	synth := controlledParticipleSynth(t,
		"gasto\tgastar\tVMP00SM\n"+
			"gastado\tgastar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)

	tok := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 7, "msg", "", "try {suggestion} / {Suggestion} / {SUGGESTION}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"try gasto / Gasto / GASTO"}, out.GetSuggestedReplacements())
}

func TestRegularIrregularParticipleFilter_NullPaths(t *testing.T) {
	synth := controlledParticipleSynth(t,
		"gasto\tgastar\tVMP00SM\n"+
			"gastado\tgastar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)

	// no VMP → null
	pos := "NCMS000"
	lem := "casa"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("casa", &pos, &lem), 0)
	m := baseMatch(0, 4, "msg", "", "{suggestion}")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil))

	// length ≤ 1 (only one form) → null
	single := controlledParticipleSynth(t, "gastado\tgastar\tVMP00SM\n")
	f.SetSynthesizer(single)
	vmp := vmpToken("gastado", "gastar", "VMP00SM", 0)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{vmp}, nil))

	// wrong direction regularity: RegularToIrregular but token is irregular → null
	f.SetSynthesizer(synth)
	irr := vmpToken("gasto", "gastar", "VMP00SM", 0)
	m2 := baseMatch(0, 5, "msg", "", "{suggestion}")
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{irr}, nil))

	// IrregularToRegular but token is regular → null
	reg := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m3 := baseMatch(0, 7, "msg", "", "{suggestion}")
	require.Nil(t, f.AcceptRuleMatch(m3, map[string]string{"direction": "IrregularToRegular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{reg}, nil))
}

func TestRegularIrregularParticipleFilter_DesiredPostagCommon(t *testing.T) {
	// endsWith "C" → prefix + "[MC]"
	// Manual tags VMP00SM and VMP00SC so regex VMP00[MC][MC] matches both after rewrite of VMP00SC
	// Actually for VMP00SC: ends with C → VMP00S + [MC] = VMP00S[MC]
	synth := controlledParticipleSynth(t,
		"assente\tassentar\tVMP00SC\n"+
			"assentado\tassentar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)

	// token irregular assente (doesn't end do/da/...) → IrregularToRegular
	tok := vmpToken("assente", "assentar", "VMP00SC", 0)
	m := baseMatch(0, 7, "msg", "", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "IrregularToRegular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"assentado"}, out.GetSuggestedReplacements())
}

func TestRegularIrregularParticipleFilter_MissingDirectionPanics(t *testing.T) {
	f := NewRegularIrregularParticipleFilter()
	m := baseMatch(0, 1, "msg", "", "{suggestion}")
	require.PanicsWithValue(t, "Missing key 'direction'", func() {
		f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
}

func TestRegularIrregularParticipleFilter_NoStartPosMatchPanics(t *testing.T) {
	// Java: atr stays null → NPE on hasPosTagStartingWith. No invent single-token fallback.
	f := NewRegularIrregularParticipleFilter()
	tok := vmpToken("gastado", "gastar", "VMP00SM", 10) // start 10 ≠ match.from 0
	m := baseMatch(0, 7, "msg", "", "{suggestion}")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
			[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}

func TestRegularIrregularParticipleFilter_FailClosedWithoutDualForms(t *testing.T) {
	// INSTANCE without dict → empty synthesize → null (do not invent forms)
	ClearDefaultParticipleSynthesizer()
	f := NewRegularIrregularParticipleFilter()
	// force empty synth
	f.SetSynthesizer(ptsynthesis.INSTANCE)
	tok := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 7, "msg", "", "{suggestion}")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil))
}

func TestRegularIrregularParticipleFilter_SynthesizeUsesSelectedTokenLemma(t *testing.T) {
	// Java synthesizes from selectedAT (last VMP reading): lemma drives forms.
	// "feita" irregular; "fazida" regular (ends with "da").
	synth := controlledParticipleSynth(t,
		"feita\tfazer\tVMP00SF\n"+
			"fazida\tfazer\tVMP00SF\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)
	tok := vmpToken("feita", "fazer", "VMP00SF", 0)
	m := baseMatch(0, 5, "msg", "", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "IrregularToRegular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fazida"}, out.GetSuggestedReplacements())
}

func TestRegularIrregularParticipleFilter_DirectionCaseInsensitive(t *testing.T) {
	synth := controlledParticipleSynth(t,
		"gasto\tgastar\tVMP00SM\n"+
			"gastado\tgastar\tVMP00SM\n",
	)
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)
	tok := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 7, "msg", "", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "regulartoirregular"}, 0,
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

// Optional integration: real portuguese_synth.dict if present in third_party.
func TestRegularIrregularParticipleFilter_RealDictOptional(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// rules/pt → repo root is 6 levels up: pt, rules, languagetool, org, languagetool, internal → root
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
	dictPath := filepath.Join(root, "third_party", "portuguese-pos-dict", "org", "languagetool", "resource", "pt", "portuguese_synth.dict")
	synth := ptsynthesis.OpenPortugueseSynthesizerFromDictPath(dictPath)
	if synth == nil {
		t.Skip("portuguese_synth.dict not loadable")
	}
	// Probe: synthesize gastar with VMP00S[MC]
	pos := "VMP00SM"
	lem := "gastar"
	at := languagetool.NewAnalyzedToken("gastado", &pos, &lem)
	forms, err := synth.SynthesizeRE(at, "VMP00S[MC]", true)
	require.NoError(t, err)
	if len(forms) <= 1 {
		t.Skipf("dict did not yield dual participles for gastar: %v", forms)
	}
	f := NewRegularIrregularParticipleFilter()
	f.SetSynthesizer(synth)
	tok := vmpToken("gastado", "gastar", "VMP00SM", 0)
	m := baseMatch(0, 7, "msg", "", "{suggestion}")
	out := f.AcceptRuleMatch(m, map[string]string{"direction": "RegularToIrregular"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	if out == nil {
		t.Skipf("filter null with forms %v (may not include irregular)", forms)
	}
	// Replacement must be a non-regular form from synthesizer
	rep := out.GetSuggestedReplacements()[0]
	require.False(t, IsRegularParticiple(rep), "replacement %q should be irregular", rep)
}
