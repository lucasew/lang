package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPattern_FlyByNight(t *testing.T) {
	toks := []*PatternToken{
		NewPatternToken("fly", false, false, false),
		NewPatternToken("by", false, false, false),
		NewPatternToken("night", false, false, false),
		func() *PatternToken {
			p := NewPatternToken("", false, false, false)
			p.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
			return p
		}(),
	}
	for _, p := range toks {
		p.SetInsideMarker(p.Token != "")
	}
	rule := NewAbstractTokenBasedRule("CA_FLY_BY_NIGHT", "t", "en", toks)
	m := NewPatternRuleMatcher(rule)
	nn := "NN"
	ats := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("These", strPtr("DT"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fly", strPtr("VB"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("by", strPtr("IN"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("night", strPtr("NN"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("companies", &nn, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(".", strPtr("PCT"), nil)),
	}
	pos := 0
	for _, a := range ats {
		a.SetStartPos(pos)
		pos += len(a.GetToken()) + 1
	}
	ms, err := m.Match(languagetool.NewAnalyzedSentence(ats))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

func TestPattern_WaitingPatient(t *testing.T) {
	wait := NewPatternToken("wait", false, false, true)
	prp := NewPatternToken("", false, false, false)
	prp.SetPosToken(PosToken{PosTag: "PRP$", Regexp: false})
	nn := NewPatternToken("", false, false, false)
	nn.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	vb := NewPatternToken("", false, false, false)
	vb.SetPosToken(PosToken{PosTag: "VB", Regexp: false})
	rule := NewAbstractTokenBasedRule("W", "t", "en", []*PatternToken{wait, prp, nn, vb})
	m := NewPatternRuleMatcher(rule)
	ats := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("waiting", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("my", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("patient", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("finish", nil, nil)),
	}
	pos := 0
	for _, a := range ats {
		a.SetStartPos(pos)
		pos += len(a.GetToken()) + 1
	}
	ms, err := m.Match(languagetool.NewAnalyzedSentence(ats))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

func TestPattern_POSExceptionBlocksMultiReading(t *testing.T) {
	// Java: pattern postag VB. matches VBZ, but exception postag NN.* on NNS reading
	// rejects the whole token (isExceptionMatchedCompletely).
	pt := NewPatternToken("", false, false, false)
	pt.SetPosToken(PosToken{PosTag: "SENT_END|VB.", Regexp: true})
	pt.SetStringPosExceptionFull("", false, false, "NN.*|VBN", true)
	m := NewPatternTokenMatcher(pt)
	nns, vbz := "NNS", "VBZ"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("companies", &nns, strPtr("company")),
		languagetool.NewAnalyzedToken("companies", &vbz, strPtr("company")),
	}, 0)
	require.False(t, m.IsMatchedReadings(atr), "POS exception on NNS should reject")

	// only VBZ → match
	atr2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("flies", &vbz, strPtr("fly")))
	require.True(t, m.IsMatchedReadings(atr2))
}
