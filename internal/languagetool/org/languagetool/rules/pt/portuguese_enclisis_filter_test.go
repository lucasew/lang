package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ptsynthesis "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	"github.com/stretchr/testify/require"
)

func enclisisATR(surface, lemma, postag string) *languagetool.AnalyzedTokenReadings {
	var posPtr *string
	var lemPtr *string
	if postag != "" {
		p := postag
		posPtr = &p
	}
	if lemma != "" {
		l := lemma
		lemPtr = &l
	}
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(surface, posPtr, lemPtr), 0)
}

func enclisisATRMulti(surface string, readings [][2]string) *languagetool.AnalyzedTokenReadings {
	// readings: [pos, lemma]
	if len(readings) == 0 {
		return languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(surface, nil, nil), 0)
	}
	pos0 := readings[0][0]
	lem0 := readings[0][1]
	var posPtr *string
	var lemPtr *string
	if pos0 != "" {
		posPtr = &pos0
	}
	if lem0 != "" {
		lemPtr = &lem0
	}
	enclisisATR := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(surface, posPtr, lemPtr), 0)
	for i := 1; i < len(readings); i++ {
		pos := readings[i][0]
		lem := readings[i][1]
		var pp *string
		var lp *string
		if pos != "" {
			pp = &pos
		}
		if lem != "" {
			lp = &lem
		}
		enclisisATR.AddReading(languagetool.NewAnalyzedToken(surface, pp, lp), "")
	}
	return enclisisATR
}

// controlledEnclisisSynth builds ManualSynthesizer-backed PortugueseSynthesizer.
// Manual line format: form\tlemma\tposTag
func controlledEnclisisSynth(t *testing.T, lines string) *ptsynthesis.PortugueseSynthesizer {
	t.Helper()
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(lines))
	require.NoError(t, err)
	return ptsynthesis.NewPortugueseSynthesizer(manual)
}

// mockEnclisisSynth returns forms as token+"-"+suffix from enclisis tag after first colon.
func mockEnclisisSynth(forms map[string][]string) synthesis.Synthesizer {
	return synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if f, ok := forms[posTag]; ok {
				return f, nil
			}
			// default: surface-lemma keyed form for inspection
			if i := strings.Index(posTag, ":"); i >= 0 {
				return []string{token.GetToken() + "-" + posTag[i+1:]}, nil
			}
			return []string{token.GetToken()}, nil
		},
	}
}

func TestPortugueseEnclisisFilter_PronounTags_NosMatrix(t *testing.T) {
	// nos special-case: always PP1CPO00; +PP3MPA00 when verb ends m/ão/õe; break
	tags := getPronounTags(enclisisATR("nos", "nós", "PP1CPO00"), "dizem", false)
	require.Equal(t, []string{"PP1CPO00", "PP3MPA00"}, tags)

	tags = getPronounTags(enclisisATR("nos", "nós", "PP1CPO00"), "farão", false)
	require.Equal(t, []string{"PP1CPO00", "PP3MPA00"}, tags)

	tags = getPronounTags(enclisisATR("nos", "nós", "PP1CPO00"), "põe", false)
	require.Equal(t, []string{"PP1CPO00", "PP3MPA00"}, tags)

	// no second tag when verb does not end m/ão/õe
	tags = getPronounTags(enclisisATR("nos", "nós", "PP1CPO00"), "diz", false)
	require.Equal(t, []string{"PP1CPO00"}, tags)

	// nos breaks even if later readings would add PP tags
	multi := enclisisATRMulti("nos", [][2]string{{"PP1CPO00", "nós"}, {"PP3MPA00", "eles"}})
	tags = getPronounTags(multi, "diz", false)
	require.Equal(t, []string{"PP1CPO00"}, tags)
}

func TestPortugueseEnclisisFilter_PronounTags_AccusativeConversion(t *testing.T) {
	// convertToAccusative: N00 → A00
	tags := getPronounTags(enclisisATR("eles", "eles", "PP3MPN00"), "ver", true)
	require.Equal(t, []string{"PP3MPA00"}, tags)

	// without conversion
	tags = getPronounTags(enclisisATR("eles", "eles", "PP3MPN00"), "ver", false)
	require.Equal(t, []string{"PP3MPN00"}, tags)

	// already accusative — no change
	tags = getPronounTags(enclisisATR("os", "eles", "PP3MPA00"), "ver", true)
	require.Equal(t, []string{"PP3MPA00"}, tags)

	// non-PP skipped
	tags = getPronounTags(enclisisATR("casa", "casa", "NCMS000"), "ver", true)
	require.Empty(t, tags)

	// null POS skipped
	tags = getPronounTags(enclisisATR("x", "x", ""), "ver", true)
	require.Empty(t, tags)

	// multiple PP readings collected
	multi := enclisisATRMulti("o", [][2]string{{"PP3MSA00", "ele"}, {"PP3MSA00", "o"}})
	tags = getPronounTags(multi, "ver", false)
	require.Equal(t, []string{"PP3MSA00", "PP3MSA00"}, tags)
}

func TestPortugueseEnclisisFilter_ConvertPronounToAccusative(t *testing.T) {
	require.Equal(t, "PP3MPA00", convertPronounToAccusative("PP3MPN00"))
	require.Equal(t, "PP3FSA00", convertPronounToAccusative("PP3FSN00"))
	require.Equal(t, "PP3MPA00", convertPronounToAccusative("PP3MPA00"))
	require.Equal(t, "XYZ", convertPronounToAccusative("XYZ"))
}

func TestPortugueseEnclisisFilter_TitleAndAllCapsCasing(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return []string{"tinham-no"}, nil
		},
	})

	// title case surface → UppercaseFirstChar on form
	verb := enclisisATR("Tinham", "ter", "VMII3P0")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	m := rules.NewRuleMatch(nil, nil, 0, 8, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Tinham-no"}, out.GetSuggestedReplacements())

	// ALLCAPS surface → toUpper
	verb = enclisisATR("TINHAM", "ter", "VMII3P0")
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 8, "msg"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.Equal(t, []string{"TINHAM-NO"}, out.GetSuggestedReplacements())

	// lowercase unchanged
	verb = enclisisATR("tinham", "ter", "VMII3P0")
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 8, "msg"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.Equal(t, []string{"tinham-no"}, out.GetSuggestedReplacements())
}

func TestPortugueseEnclisisFilter_AcceptRuleMatch_HappyPath(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(mockEnclisisSynth(nil))

	verb := enclisisATR("tinham", "ter", "VMIP3P0")
	mid := enclisisATR("-", "", "")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	m := rules.NewRuleMatch(nil, nil, 0, 8, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"verbPos": "0", "pronounPos": "2", "convertToAccusative": "False",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, mid, pron}, nil)
	require.NotNil(t, out)
	require.Same(t, m, out)
	require.Equal(t, []string{"tinham-PP3MSA00"}, out.GetSuggestedReplacements())
}

func TestPortugueseEnclisisFilter_AcceptRuleMatch_EmptyTagsNull(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(mockEnclisisSynth(nil))

	verb := enclisisATR("tinham", "ter", "VMIP3P0")
	// no PP tags → null
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 8, "msg"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, enclisisATR("xyz", "xyz", "NCMS000")}, nil)
	require.Nil(t, out)

	// null POS → empty tags → null
	out = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 8, "msg"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, enclisisATR("xyz", "xyz", "")}, nil)
	require.Nil(t, out)
}

func TestPortugueseEnclisisFilter_AcceptRuleMatch_ConvertToAccusativeParse(t *testing.T) {
	// Boolean.parseBoolean: only "true" (ignore case) is true
	f := NewPortugueseEnclisisFilter()
	// Capture the enclisis tag requested
	var gotTag string
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			gotTag = posTag
			return []string{"ver-os"}, nil
		},
	})
	verb := enclisisATR("ver", "ver", "VMN0000")
	pron := enclisisATR("eles", "eles", "PP3MPN00")

	// "true" → convert
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 3, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "true",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Equal(t, "VMN0000:PP3MPA00", gotTag)

	// "TRUE" → convert
	_ = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 3, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "TRUE",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.Equal(t, "VMN0000:PP3MPA00", gotTag)

	// "False" / missing → no convert (Java Boolean.parseBoolean)
	_ = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 3, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "False",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.Equal(t, "VMN0000:PP3MPN00", gotTag)

	_ = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 3, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.Equal(t, "VMN0000:PP3MPN00", gotTag)
}

func TestPortugueseEnclisisFilter_AcceptRuleMatch_InvalidPosFailClosed(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(mockEnclisisSynth(nil))
	verb := enclisisATR("ver", "ver", "VMN0000")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	// bad parse
	require.Nil(t, f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 1, "m"), map[string]string{
		"verbPos": "x", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil))
	require.Nil(t, f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 1, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "y", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil))
	// OOB fail-closed
	require.Nil(t, f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 1, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "5", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil))
}

func TestPortugueseEnclisisFilter_FirstVReadingOnly(t *testing.T) {
	// break after first V* reading
	f := NewPortugueseEnclisisFilter()
	var synthCalls int
	var usedPOS string
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			synthCalls++
			if token.GetPOSTag() != nil {
				usedPOS = *token.GetPOSTag()
			}
			return []string{"form"}, nil
		},
	})
	// first reading non-V, second V, third V — only second used, break
	verb := enclisisATRMulti("tinham", [][2]string{
		{"NCMS000", "tinham"},
		{"VMII3P0", "ter"},
		{"VMIP3P0", "ter"},
	})
	pron := enclisisATR("o", "ele", "PP3MSA00")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 6, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Equal(t, 1, synthCalls)
	require.Equal(t, "VMII3P0", usedPOS)
	require.Equal(t, []string{"form"}, out.GetSuggestedReplacements())
}

func TestPortugueseEnclisisFilter_UsesFullAnalyzedTokenForSynth(t *testing.T) {
	// Manual synth keyed by lemma+enclisis tag — proves full AT (lemma) is passed.
	manual := controlledEnclisisSynth(t, "tinham-no\tter\tVMIP3P0:PP3MSA00\n")
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(manual)
	verb := enclisisATR("tinham", "ter", "VMIP3P0")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 6, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "tinham-no")
}

func TestPortugueseEnclisisFilter_NosPathSynthesizesBothTags(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			switch {
			case strings.HasSuffix(posTag, ":PP1CPO00"):
				return []string{"dizem-nos"}, nil
			case strings.HasSuffix(posTag, ":PP3MPA00"):
				return []string{"dizem-os"}, nil
			default:
				return nil, nil
			}
		},
	})
	verb := enclisisATR("dizem", "dizer", "VMIP3P0")
	pron := enclisisATR("nos", "nós", "PP1CPO00")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 8, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	got := out.GetSuggestedReplacements()
	// HashSet membership — not order
	require.Contains(t, got, "dizem-nos")
	require.Contains(t, got, "dizem-os")
	require.Len(t, got, 2)
}

func TestPortugueseEnclisisFilter_DefaultSynthesizerIsINSTANCE(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	require.Same(t, ptsynthesis.INSTANCE, f.getSynthesizer())
}

func TestPortugueseEnclisisFilter_DefaultINSTANCEPath(t *testing.T) {
	// Default path uses INSTANCE (dict optional). With empty/unloaded INSTANCE,
	// synthesize may return empty; still exercises AcceptRuleMatch wiring.
	f := NewPortugueseEnclisisFilter()
	require.Same(t, ptsynthesis.INSTANCE, f.getSynthesizer())

	verb := enclisisATR("tinham", "ter", "VMIP3P0")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 6, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	// pronoun tags present → match returned with whatever INSTANCE yields (often empty without dict)
	require.NotNil(t, out)
	// mock path covered elsewhere; if INSTANCE has forms, fine
	_ = out.GetSuggestedReplacements()
}

func TestPortugueseEnclisisFilter_EmptySynthYieldsEmptySuggestions(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SetSynthesizer(synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return nil, nil
		},
	})
	verb := enclisisATR("tinham", "ter", "VMIP3P0")
	pron := enclisisATR("o", "ele", "PP3MSA00")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 6, "m"), map[string]string{
		"verbPos": "0", "pronounPos": "1", "convertToAccusative": "false",
	}, 0, []*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Empty(t, out.GetSuggestedReplacements())
}

func TestPortugueseEnclisisFilter_NilFilterOrMatch(t *testing.T) {
	var f *PortugueseEnclisisFilter
	require.Nil(t, f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 1, "m"), nil, 0, nil, nil))
	require.Nil(t, NewPortugueseEnclisisFilter().AcceptRuleMatch(nil, map[string]string{
		"verbPos": "0", "pronounPos": "0",
	}, 0, nil, nil))
}

func TestPortugueseEnclisisFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.pt.PortugueseEnclisisFilter"))
	created := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.pt.PortugueseEnclisisFilter")
	require.NotNil(t, created)
	pf, ok := created.(*PortugueseEnclisisFilter)
	require.True(t, ok)
	require.Same(t, ptsynthesis.INSTANCE, pf.getSynthesizer())
}
