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
