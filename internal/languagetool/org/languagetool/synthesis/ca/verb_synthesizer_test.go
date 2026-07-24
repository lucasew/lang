package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func ptrVS(s string) *string { return &s }

func atrVS(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken(token, ptrVS(pos), ptrVS(lemma)))
}

func TestVerbSynthesizer_FindVerbGroup(t *testing.T) {
	vtag := "VMIP3S0"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Ell", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("menja", &vtag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("pa", nil, nil)),
	}
	v := NewVerbSynthesizer(tokens)
	require.True(t, v.FindVerbGroup())
	require.Equal(t, 1, v.IFirstVerb)
	require.Equal(t, 1, v.ILastVerb)
	v.SetTarget("menjar", "VMN0000")
	require.True(t, v.HasTarget())
}

func TestVerbSynthesizer_IndexesSimple(t *testing.T) {
	// [0]SENT-like noun, [1]verb
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("Ell", "PP3MS000", "ell"),
		atrVS("menja", "VMIP3S00", "menjar"),
		atrVS("pa", "NCMS000", "pa"),
	}
	v := NewVerbSynthesizerAt(tokens, 1, false)
	require.False(t, v.IsUndefined())
	require.Equal(t, 1, v.GetFirstVerbIndex())
	require.Equal(t, 1, v.GetLastVerbIndex())
	require.Equal(t, 0, v.GetNumPronounsBefore())
	require.Equal(t, 0, v.GetNumPronounsAfter())
	require.Equal(t, "menja", v.GetVerbStr())
	require.Equal(t, "3S", v.GetFirstVerbPersonaNumber())
	require.True(t, v.IsFirstVerbIS())
	require.Equal(t, "VMIP3S00", v.GetFirstVerbISPostag())
}

func TestVerbSynthesizer_PronounsAfter(t *testing.T) {
	// menjar-ho: no whitespace before clitic
	verb := atrVS("menjar", "VMN00000", "menjar")
	clitic := atrVS("ho", "PP3NNA00", "ho")
	// PP3NNA00 matches PP3..A00
	clitic.SetWhitespaceBefore(false)
	tokens := []*languagetool.AnalyzedTokenReadings{verb, clitic}
	v := NewVerbSynthesizerAt(tokens, 0, false)
	require.False(t, v.IsUndefined())
	require.Equal(t, 1, v.GetNumPronounsAfter())
	require.Equal(t, "ho", v.GetPronounsStrAfter())
	require.Equal(t, 1, v.GetLastIndex())
	require.Equal(t, "menjarho", v.GetWholeOriginalStr()) // no space when !IsWhitespaceBefore
}

func TestVerbSynthesizer_PronounsBefore(t *testing.T) {
	// em menja
	pr := atrVS("em", "PP1CS000", "jo")
	// PP1CS000 — does NOT match pPronomFeble (only PP[123]CP000, PP3CSD00, P0, etc.)
	// Use P0 form or PP3CSD00 for dative li, or PP1CP000 for ens-like
	// pPronomFeble: P0.{6}|PP3CN000|PP3NN000|PP3..A00|PP[123]CP000|PP3CSD00
	// "em" is often P0xxxxx — e.g. P01CN000
	pr = atrVS("em", "P01CN000", "jo")
	pr.SetWhitespaceBefore(true)
	verb := atrVS("menja", "VMIP3S00", "menjar")
	verb.SetWhitespaceBefore(true)
	// Need index 0 something so iFirstVerb+i > 0 loop works: tokens[0] dummy, [1]em, [2]menja
	// Java: while (iFirstVerb + i > 0 && ...) — when iFirstVerb=1, i=-1 → index 0 must not be checked with >0?
	// iFirstVerb=1, i=-1: 1-1=0, condition iFirstVerb+i > 0 is false (0>0). So pronouns before from index 0 never counted!
	// When iFirstVerb=2 (em at 1, verb at 2): i=-1 → index 1 > 0, counts em if pPronomFeble.
	dummy := atrVS("X", "NCFS000", "x")
	tokens := []*languagetool.AnalyzedTokenReadings{dummy, pr, verb}
	v := NewVerbSynthesizerAt(tokens, 2, false)
	require.False(t, v.IsUndefined())
	require.Equal(t, 2, v.GetFirstVerbIndex())
	require.Equal(t, 1, v.GetNumPronounsBefore())
	require.Equal(t, "em", v.GetPronounsStrBefore())
	require.Equal(t, "em menja", v.GetWholeOriginalStr())
	require.Equal(t, "em menja", v.GetCasingModel()) // from first-num to first verb inclusive
}

func TestVerbSynthesizer_SynthesizeForm(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("menja", "VMIP3S00", "menjar"),
	}
	v := NewVerbSynthesizerAt(tokens, 0, false)
	v.SetPostag("VMIP1S00")
	require.Equal(t, "menjar", v.NewLemma)
	require.Equal(t, "VMIP1S00", v.NewPostag)
	v.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		if postag == "VMIP1S00" {
			return []string{"menjo"}
		}
		return nil
	}
	require.Equal(t, "menjo", v.SynthesizeForm())
}

func TestVerbSynthesizer_AdjustHaverSer(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("ha", "VAIP3S00", "haver"),
	}
	v := NewVerbSynthesizerAt(tokens, 0, false)
	v.SetLemmaAndPostag("haver", "VMIP1S00")
	v.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		// adjustPostagToLemma → VAIP1S00
		require.Equal(t, "VAIP1S00", postag)
		return []string{"he"}
	}
	require.Equal(t, "he", v.SynthesizeForm())
}

func TestVerbSynthesizer_SingleParticiple(t *testing.T) {
	// participle without GV → single-token group, no pronoun scan expansion
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("oblidat", "VMP00SM0", "oblidar"),
	}
	v := NewVerbSynthesizerAt(tokens, 0, false)
	require.False(t, v.IsUndefined())
	require.Equal(t, 0, v.GetFirstVerbIndex())
	require.Equal(t, 0, v.GetNumPronounsAfter())
	require.Equal(t, 0, v.GetNumPronounsBefore())
}

func TestVerbSynthesizer_PassatPerifrastic(t *testing.T) {
	// needs iFirstVerb >= 1
	dummy := atrVS("ell", "PP3MS000", "ell")
	va := atrVS("va", "VAIP3S00", "anar")
	va.SetChunkTags([]string{"GV"})
	inf := atrVS("menjar", "VMN00000", "menjar")
	inf.SetChunkTags([]string{"GV"})
	tokens := []*languagetool.AnalyzedTokenReadings{dummy, va, inf}
	v := NewVerbSynthesizerAt(tokens, 1, false)
	require.True(t, v.IsPassatPerifrastic())
	require.False(t, v.IsPerfet())
}

func TestVerbSynthesizer_SearchBackward(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("menja", "VMIP3S00", "menjar"),
		atrVS("pa", "NCMS000", "pa"),
		atrVS(".", "_PUNCT", "."),
	}
	// start at end, search backward to verb
	v := NewVerbSynthesizerAt(tokens, 2, true)
	require.False(t, v.IsUndefined())
	require.Equal(t, 0, v.GetFirstVerbIndex())
}

func TestVerbSynthesizer_Undefined(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrVS("pa", "NCMS000", "pa"),
	}
	v := NewVerbSynthesizerAt(tokens, 0, false)
	require.True(t, v.IsUndefined())
}
