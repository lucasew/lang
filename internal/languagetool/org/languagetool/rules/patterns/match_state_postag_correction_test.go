package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

type correctingSynth struct {
	synthesis.FuncSynthesizer
}

func (correctingSynth) GetPosTagCorrection(posTag string) string {
	return "CORR:" + posTag
}

func TestMatchState_GetTargetPosTag_PosTagCorrectionWhenSetPos(t *testing.T) {
	// postag regexp + replace + setpos → correction applied (non-static lemma path)
	m := NewMatch("NN", "VB", true, "", "", CaseNone, true, false, IncludeNone)
	// PosTag is the regexp when postagRegexp; NewMatch with postagRegexp true compiles PosTag as pos regex
	// Wait: NewMatch(posTag, posTagReplace, postagRegexp, ...)
	// So posTag="NN" as regex matches "NN", replace with "VB"
	pos := "NN"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("run", &pos, nil),
	}, 0)
	st := NewMatchStateWithSynth(m, correctingSynth{})
	st.SetToken(tok)
	got := st.GetTargetPosTag()
	require.Equal(t, "CORR:VB", got)

	// without setpos, no correction
	m2 := NewMatch("NN", "VB", true, "", "", CaseNone, false, false, IncludeNone)
	st2 := NewMatchStateWithSynth(m2, correctingSynth{})
	st2.SetToken(tok)
	require.Equal(t, "VB", st2.GetTargetPosTag())
}

func TestMatchState_GetTargetPosTag_FullMatchPOS(t *testing.T) {
	// VB.? should not match VBGX via substring — full match only
	m := NewMatch("VB.?", "$0", true, "", "", CaseNone, false, false, IncludeNone)
	pos := "VBGX"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("x", &pos, nil),
	}, 0)
	st := NewMatchState(m)
	st.SetToken(tok)
	// no full match → empty posTags → target stays template; replace path with empty adds targetPosTag
	// GetTargetPosTag with replace and empty posTags adds original target "VB.?" then replace → still "VB.?"
	got := st.GetTargetPosTag()
	require.NotContains(t, got, "VBGX")
}

// Twin of getTargetPosTag: postag_replace only when attr present (Java != null).
func TestMatchState_GetTargetPosTag_PostagReplacePresent(t *testing.T) {
	pos := "N:f:sg:acc"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("x", &pos, nil),
	}, 0)

	// postag regexp without replace attr → keep matched tag as target (no invent replace)
	mNo := NewMatch("N:.*", "", true, "", "", CaseNone, false, false, IncludeNone)
	require.False(t, mNo.PosTagReplacePresent)
	stNo := NewMatchState(mNo)
	stNo.SetToken(tok)
	require.Equal(t, "N:f:sg:acc", stNo.GetTargetPosTag()) // last matched / Base last

	// with replace
	mYes := NewMatch("N:([fm]):(sg|pl):(acc|nom)", "N:$1:$2:$3", true, "", "", CaseNone, true, false, IncludeNone)
	require.True(t, mYes.PosTagReplacePresent)
	stYes := NewMatchState(mYes)
	stYes.SetToken(tok)
	require.Equal(t, "N:f:sg:acc", stYes.GetTargetPosTag())
}

// Twin of BaseSynthesizer.getTargetPosTag: last matching POS when synth has no override.
func TestMatchState_GetTargetPosTag_LastTagFallback(t *testing.T) {
	// postag regexp matching both NN and NNS; no replace → pick last via Base fallback
	m := NewMatch("NN.*", "", true, "", "", CaseNone, false, false, IncludeNone)
	// Two readings: NN then NNS
	nn, nns := "NN", "NNS"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("dogs", &nn, nil),
		languagetool.NewAnalyzedToken("dogs", &nns, nil),
	}, 0)
	// FuncSynthesizer has no GetTargetPosTag → last-tag Base fallback
	st := NewMatchStateWithSynth(m, synthesis.FuncSynthesizer{})
	st.SetToken(tok)
	require.Equal(t, "NNS", st.GetTargetPosTag())

	// BaseSynthesizer path (explicit) same last tag
	st2 := NewMatchStateWithSynth(m, synthesis.NewBaseSynthesizer("en", nil))
	st2.SetToken(tok)
	require.Equal(t, "NNS", st2.GetTargetPosTag())
}
