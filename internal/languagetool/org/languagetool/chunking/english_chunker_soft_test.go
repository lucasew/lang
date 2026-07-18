package chunking

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishChunker_DTRestartsNP(t *testing.T) {
	// "his chair an" → two NPs so PAST_AN_PAST can match E-NP then "an"
	his, chair, an := "PRP$", "NN", "DT"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("his", &his, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("chair", &chair, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("an", &an, nil)),
	}
	NewEnglishChunker().AddChunkTags(toks)
	require.Contains(t, strings.Join(toks[1].GetChunkTags(), ","), "E-NP")
	require.Contains(t, strings.Join(toks[2].GetChunkTags(), ","), "B-NP")
}

func TestEnglishChunker_ToTellVP(t *testing.T) {
	toTag, nn, vb := "IN", "NN", "VB"
	to := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("to", &toTag, nil),
		languagetool.NewAnalyzedToken("to", strPtr("TO"), nil),
	}, 0)
	tell := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("tell", &nn, nil),
		languagetool.NewAnalyzedToken("tell", &vb, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{to, tell})
	require.Contains(t, strings.Join(tell.GetChunkTags(), ","), "VP")
}

func strPtr(s string) *string { return &s }

func TestEnglishChunkFilter_KnowsDoesNotPluralizeAnyone(t *testing.T) {
	// Regression: getChunkType must not scan past the NP into NNS|VBZ "knows".
	nns, vbz, prp := "NNS", "VBZ", "PRP"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("Does", &nns, strPtr("do")),
			languagetool.NewAnalyzedToken("Does", &vbz, strPtr("do")),
		}, 0),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("anyone", &prp, strPtr("anyone"))),
		languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
			languagetool.NewAnalyzedToken("knows", &nns, strPtr("know")),
			languagetool.NewAnalyzedToken("knows", &vbz, strPtr("know")),
		}, 0),
	}
	NewEnglishChunker().AddChunkTags(toks)
	require.Contains(t, strings.Join(toks[1].GetChunkTags(), ","), "singular")
	require.Contains(t, strings.Join(toks[0].GetChunkTags(), ","), "VP")
}
