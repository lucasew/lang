package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestCatalanSynthesizer_LemmasToIgnore(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader("x\tsentar\tVMIP3S00\n"))
	require.NoError(t, err)
	s := NewCatalanSynthesizer(man)
	lemma := "sentar"
	pos := "VMIP3S00"
	tok := languagetool.NewAnalyzedToken("senta", &pos, &lemma)
	// SynthesizeRE with regexp path ignores lemma
	forms, err := s.SynthesizeRE(tok, "VMIP3S00", true)
	require.NoError(t, err)
	require.Empty(t, forms)
}

func TestCatalanSynthesizer_VerbLemmaWithSpace(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"fa\tfer\tVMIP3S00\n",
	))
	require.NoError(t, err)
	s := NewCatalanSynthesizer(man)
	// possibleTags from manual
	lemma := "fer fred"
	pos := "VMIP3S00"
	tok := languagetool.NewAnalyzedToken("fa", &pos, &lemma)
	// Java synthesize(token, posTag) treats posTag as Pattern against possibleTags
	forms, err := s.Synthesize(tok, pos)
	require.NoError(t, err)
	require.Equal(t, []string{"fa fred"}, forms)
}

func TestCatalanSynthesizer_GetTargetPosTag(t *testing.T) {
	s := NewCatalanSynthesizer(nil)
	require.Equal(t, "X", s.GetTargetPosTag(nil, "X"))
	// Indicative (I at pos 2) sorts after non-I; last = indicative
	// charAt(2)=='I' gets +100 vs non-I
	got := s.GetTargetPosTag([]string{"VMSP3S00", "VMIP3S00"}, "fb")
	require.Equal(t, "VMIP3S00", got)
	// 3 person > 1 person at char 4
	got2 := s.GetTargetPosTag([]string{"VMIP1S00", "VMIP3S00"}, "fb")
	require.Equal(t, "VMIP3S00", got2)
	// VMIP2P00 > VMIS3S00
	got3 := s.GetTargetPosTag([]string{"VMIS3S00", "VMIP2P00"}, "fb")
	require.Equal(t, "VMIP2P00", got3)
}

func TestCatalanSynthesizer_RegionalInstances(t *testing.T) {
	require.Equal(t, "ca-ES", INSTANCE_CAT.LanguageCode)
	require.Equal(t, "ca-ES-valencia", INSTANCE_VAL.LanguageCode)
	require.Equal(t, "ca-ES-balear", INSTANCE_BAL.LanguageCode)
	require.Equal(t, "[0CXY12]", INSTANCE_CAT.verbTagSuffix())
	require.Equal(t, "[0VXZ13567]", INSTANCE_VAL.verbTagSuffix())
	require.Equal(t, "[0BYZ1247]", INSTANCE_BAL.verbTagSuffix())
}

func TestCatalanSynthesizer_RegionalFallbackPattern(t *testing.T) {
	// When exact tag miss, regional suffix replaces last char of pattern.
	// Manual only has Valencia-style tag ending with V in last position.
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"parle\tparlar\tVMIP1S0V\n",
	))
	require.NoError(t, err)
	s := NewCatalanSynthesizerForLang(man, "ca-ES-valencia")
	lemma := "parlar"
	// Request ending with 0 (Catalan common); empty → fallback posTag[:-1]+[0VXZ…]
	// Pattern VMIP1S0[0VXZ13567] should match VMIP1S0V
	pos := "VMIP1S00"
	tok := languagetool.NewAnalyzedToken("parle", &pos, &lemma)
	forms, err := s.Synthesize(tok, pos)
	require.NoError(t, err)
	require.Equal(t, []string{"parle"}, forms)
}

func TestOpenCatalanSynthesizer_Missing(t *testing.T) {
	require.Nil(t, OpenCatalanSynthesizerFromDir("", "ca-ES"))
	require.Nil(t, OpenCatalanSynthesizerFromDictPath("", "ca-ES"))
}
