package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGRPCUtilsRoundTrip(t *testing.T) {
	pos := "NN"
	lem := "dog"
	tok := languagetool.NewAnalyzedToken("dogs", &pos, &lem)
	g := TokenToGRPC(tok)
	require.Equal(t, "dogs", g.Token)
	require.Equal(t, "NN", g.PosTag)
	back := TokenFromGRPC(g)
	require.Equal(t, "dogs", back.GetToken())
	require.Equal(t, "NN", *back.GetPOSTag())

	rd := languagetool.NewAnalyzedTokenReadingsAt(tok, 3)
	rd.SetChunkTags([]string{"B-NP"})
	gr := ReadingsToGRPC(rd)
	require.Equal(t, 3, gr.StartPos)
	require.Equal(t, []string{"B-NP"}, gr.ChunkTags)

	sent := languagetool.AnalyzePlain("dogs bark")
	gs := SentenceToGRPC(sent)
	require.Equal(t, "dogs bark", gs.Text)
	require.NotEmpty(t, gs.Tokens)

	s2 := SentenceFromGRPC(gs)
	require.Equal(t, gs.Text, s2.GetText())

	m := NewRuleMatch(NewFakeRule("X"), sent, 0, 4, "msg")
	m.SetSuggestedReplacements([]string{"dog"})
	gm := MatchToGRPC(m)
	require.Equal(t, "X", gm.ID)
	require.Equal(t, 0, gm.Offset)
	require.Equal(t, 4, gm.Length)

	require.Equal(t, "a b", NormalizeWhitespaceForGRPC("a\u00a0b"))
}
