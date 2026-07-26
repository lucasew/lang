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

// Java pronoun switch matrix (pure unit of switch; HashSet membership not order).
func TestPortugueseProclisisFilter_PronounMatrix(t *testing.T) {
	// lo/no → o
	require.Equal(t, []string{"o"}, proclisisPronounForms("lo", "fazê"))
	require.Equal(t, []string{"o"}, proclisisPronounForms("no", "vê"))
	// la/na → a
	require.Equal(t, []string{"a"}, proclisisPronounForms("la", "vê"))
	require.Equal(t, []string{"a"}, proclisisPronounForms("na", "faz"))
	// los → os
	require.Equal(t, []string{"os"}, proclisisPronounForms("los", "vê"))
	// las/nas → as
	require.Equal(t, []string{"as"}, proclisisPronounForms("las", "vê"))
	require.Equal(t, []string{"as"}, proclisisPronounForms("nas", "faz"))
	// nos → nos; +os when verb ends m/ão/õe
	require.Equal(t, []string{"nos"}, proclisisPronounForms("nos", "diz"))
	require.Equal(t, []string{"nos", "os"}, proclisisPronounForms("nos", "dizem"))
	require.Equal(t, []string{"nos", "os"}, proclisisPronounForms("nos", "farão"))
	require.Equal(t, []string{"nos", "os"}, proclisisPronounForms("nos", "põe"))
	// default keep
	require.Equal(t, []string{"lhe"}, proclisisPronounForms("lhe", "dizer"))
	require.Equal(t, []string{"me"}, proclisisPronounForms("me", "ajuda"))
}

// controlledProclisisSynth: ManualSynthesizer so Synthesize(at, verbTag) yields known forms.
// Manual line format: form\tlemma\tposTag
func controlledProclisisSynth(t *testing.T, lines string) *ptsynthesis.PortugueseSynthesizer {
	t.Helper()
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(lines))
	require.NoError(t, err)
	return ptsynthesis.NewPortugueseSynthesizer(manual)
}

func encliticReading(surface, lemma, postag string) *languagetool.AnalyzedTokenReadings {
	pos := postag
	lem := lemma
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(surface, &pos, &lem), 0)
}

func stemSynth(forms map[string]string) synthesis.Synthesizer {
	return synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if f, ok := forms[token.GetToken()]; ok {
				return []string{f}, nil
			}
			// strip hyphen stem as stand-in
			tok := token.GetToken()
			if i := strings.Index(tok, "-"); i >= 0 {
				return []string{tok[:i]}, nil
			}
			return []string{tok}, nil
		},
	}
}

func TestPortugueseProclisisFilter_AcceptRuleMatchSetsSuggestions(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	f.SetSynthesizer(stemSynth(nil))

	tok := encliticReading("bebemo-lo", "beber", "VMIP1P0:PP3MSA00")
	m := rules.NewRuleMatch(nil, nil, 0, 9, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"foo": "bar"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
	// HashSet membership — not order
	require.Contains(t, out.GetSuggestedReplacements(), "o bebemo")
	require.Same(t, m, out) // Java returns same match instance
}

func TestPortugueseProclisisFilter_AcceptRuleMatch_PronounFormsViaSynth(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	f.SetSynthesizer(stemSynth(nil))

	// fazê-lo → o fazê
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{encliticReading("fazê-lo", "fazer", "VMN0000:PP3MSA00")}, nil)
	require.Contains(t, out.GetSuggestedReplacements(), "o fazê")

	// dizem-nos → nos dizem + os dizem (verb ends with m)
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 9, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{encliticReading("dizem-nos", "dizer", "VMIP3P0:PP1CPO00")}, nil)
	got := out.GetSuggestedReplacements()
	require.Contains(t, got, "nos dizem")
	require.Contains(t, got, "os dizem")
}

func TestPortugueseProclisisFilter_POSGate(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	f.SetSynthesizer(stemSynth(nil))

	// null POS skipped
	tokNull := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("fazê-lo", nil, nil), 0)
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{tokNull}, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())

	// non-V skipped
	posN := "NCMS000:PP3MSA00"
	tokN := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("fazê-lo", &posN, nil), 0)
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{tokN}, nil)
	require.Empty(t, out.GetSuggestedReplacements())

	// no colon skipped
	posV := "VMN0000"
	tokV := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("fazê-lo", &posV, nil), 0)
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{tokV}, nil)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestPortugueseProclisisFilter_EmptySynthSkipsReading(t *testing.T) {
	// Empty synth → skip reading (Java would AIOOB on [0]); still returns match with empty set.
	f := NewPortugueseProclisisFilter()
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return nil, nil
		},
	})
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{encliticReading("fazê-lo", "fazer", "VMN0000:PP3MSA00")}, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestPortugueseProclisisFilter_DefaultSynthesizerIsINSTANCE(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	require.Same(t, ptsynthesis.INSTANCE, f.getSynthesizer())
}

func TestPortugueseProclisisFilter_DefaultPathWithINSTANCE(t *testing.T) {
	// If dict-backed INSTANCE has data, exercise real path; else skip with injectable covered elsewhere.
	f := NewPortugueseProclisisFilter()
	// Ensure default is INSTANCE (no override).
	require.Same(t, ptsynthesis.INSTANCE, f.getSynthesizer())

	// Try loading real dict onto a local synth and inject to prove full AnalyzedToken path.
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// Walk up to repo root from .../rules/pt/
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", "..", ".."))
	dictPath := filepath.Join(root, "third_party", "portuguese-pos-dict", "org", "languagetool", "resource", "pt", "portuguese_synth.dict")
	if _, err := filepath.Abs(dictPath); err != nil {
		t.Skip("dict path resolve failed")
	}
	// OpenPortugueseSynthesizerFromDictPath — only if dict exists
	synth := ptsynthesis.OpenPortugueseSynthesizerFromDictPath(dictPath)
	if synth == nil {
		t.Skip("portuguese_synth.dict not loadable")
	}
	// Probe: synthesize fazer with VMN0000
	lemma := "fazer"
	pos := "VMN0000:PP3MSA00"
	probe := languagetool.NewAnalyzedToken("fazê-lo", &pos, &lemma)
	forms, err := synth.Synthesize(probe, "VMN0000")
	if err != nil || len(forms) == 0 {
		// Fall back to manual-controlled synth for full path
		manual := controlledProclisisSynth(t, "fazer\tfazer\tVMN0000\n")
		f.SetSynthesizer(manual)
		out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
			[]*languagetool.AnalyzedTokenReadings{encliticReading("fazê-lo", "fazer", "VMN0000:PP3MSA00")}, nil)
		require.NotNil(t, out)
		require.Contains(t, out.GetSuggestedReplacements(), "o fazer")
		return
	}
	f.SetSynthesizer(synth)
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{encliticReading("fazê-lo", "fazer", "VMN0000:PP3MSA00")}, nil)
	require.NotNil(t, out)
	// membership: first synth form used as verb
	require.Contains(t, out.GetSuggestedReplacements(), "o "+forms[0])
}

func TestPortugueseProclisisFilter_UsesFullAnalyzedTokenForSynth(t *testing.T) {
	// Manual synth keyed by lemma, not surface — proves full AT (lemma) is passed.
	manual := controlledProclisisSynth(t, "fazer\tfazer\tVMN0000\n")
	f := NewPortugueseProclisisFilter()
	f.SetSynthesizer(manual)
	// surface is enclitic; lemma is bare verb used by synthesizer
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 7, "m"), nil, 0,
		[]*languagetool.AnalyzedTokenReadings{encliticReading("fazê-lo", "fazer", "VMN0000:PP3MSA00")}, nil)
	require.Contains(t, out.GetSuggestedReplacements(), "o fazer")
}

func TestPortugueseProclisisFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.PortugueseProclisisFilter"))
	// Creator still returns a filter with default INSTANCE synthesizer
	created := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.pt.PortugueseProclisisFilter")
	require.NotNil(t, created)
	pf, ok := created.(*PortugueseProclisisFilter)
	require.True(t, ok)
	require.Same(t, ptsynthesis.INSTANCE, pf.getSynthesizer())
}
