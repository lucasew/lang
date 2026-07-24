package es

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

// Twin of SpanishSynthesizer.synthesize verb lemma with trailing noun phrase.
func TestSpanishSynthesizer_VerbLemmaWithSpace(t *testing.T) {
	// Manual: form for first lemma word only
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"hace\thacer\tVMIP3S0\n",
	))
	require.NoError(t, err)
	s := NewSpanishSynthesizer(man)
	lemma := "hacer frio"
	pos := "VMIP3S0"
	tok := languagetool.NewAnalyzedToken("hace", &pos, &lemma)
	forms, err := s.Synthesize(tok, pos)
	require.NoError(t, err)
	require.Equal(t, []string{"hace frio"}, forms)

	// Non-verb POS: do not split lemma
	posN := "NCMS000"
	man2, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"x\thacer frio\tNCMS000\n",
	))
	require.NoError(t, err)
	s2 := NewSpanishSynthesizer(man2)
	tok2 := languagetool.NewAnalyzedToken("x", &posN, &lemma)
	forms2, err := s2.Synthesize(tok2, posN)
	require.NoError(t, err)
	require.Equal(t, []string{"x"}, forms2)
}

// Twin of SpanishSynthesizer.getTargetPosTag PostagComparator (Indicative > Imperative).
func TestSpanishSynthesizer_GetTargetPosTag(t *testing.T) {
	s := NewSpanishSynthesizer(nil)
	require.Equal(t, "X", s.GetTargetPosTag(nil, "X"))
	// Sorted ascending: VMM then VMIP; last = VMIP3S0
	got := s.GetTargetPosTag([]string{"VMM02S0", "VMIP3S0"}, "fallback")
	require.Equal(t, "VMIP3S0", got)
	got2 := s.GetTargetPosTag([]string{"VMIP3S0", "VMM02S0"}, "fallback")
	require.Equal(t, "VMIP3S0", got2)
}

func TestSpanishSynthesizer_ResourceFilenames(t *testing.T) {
	s := NewSpanishSynthesizer(nil)
	require.Equal(t, "/es/es-ES_synth.dict", s.ResourceFileName)
	require.Equal(t, "/es/es-ES_tags.txt", s.TagFileName)
	require.Equal(t, "/es/es.sor", s.SorFileName)
}

func TestOpenSpanishSynthesizer_Missing(t *testing.T) {
	require.Nil(t, OpenSpanishSynthesizerFromDir(""))
	require.Nil(t, OpenSpanishSynthesizerFromDictPath(""))
}
