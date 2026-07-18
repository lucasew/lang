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

func TestEnglishChunker_AbleThinkAndSimilarLike(t *testing.T) {
	// able think → think is VP; similar like mine → like PP, mine NP
	jj, vb, in, nn := "JJ", "VB", "IN", "NN"
	able := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("able", &jj, nil))
	think := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("think", &nn, nil),
		languagetool.NewAnalyzedToken("think", &vb, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{able, think})
	require.Contains(t, strings.Join(think.GetChunkTags(), ","), "VP")

	sim := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("similar", &jj, nil))
	like := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("like", &in, nil),
		languagetool.NewAnalyzedToken("like", &vb, nil),
	}, 0)
	mine := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("mine", &nn, nil),
		languagetool.NewAnalyzedToken("mine", &vb, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{sim, like, mine})
	require.Contains(t, strings.Join(like.GetChunkTags(), ","), "PP")
	require.Contains(t, strings.Join(mine.GetChunkTags(), ","), "NP")
}

func TestEnglishChunker_AmassingADJP(t *testing.T) {
	rb, vbg := "RB", "VBG"
	so := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("so", &rb, nil))
	am := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("amassing", &vbg, nil))
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{so, am})
	require.Contains(t, strings.Join(am.GetChunkTags(), ","), "ADJP")
}

func TestEnglishChunker_CreamColoredPaint(t *testing.T) {
	dt, jj, nn, vbd := "DT", "JJ", "NN", "VBD"
	the := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("the", &dt, nil))
	cream := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("cream", &jj, nil),
		languagetool.NewAnalyzedToken("cream", &nn, nil),
	}, 0)
	colored := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("colored", &jj, nil),
		languagetool.NewAnalyzedToken("colored", &vbd, nil),
	}, 0)
	paint := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("paint", &nn, nil))
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{the, cream, colored, paint})
	require.Contains(t, strings.Join(colored.GetChunkTags(), ","), "NP")
	require.Contains(t, strings.Join(paint.GetChunkTags(), ","), "NP")
}

func TestEnglishChunker_NotAvailableIADJP(t *testing.T) {
	rb, jj := "RB", "JJ"
	not := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("not", &rb, nil))
	av := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("available", &jj, nil))
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{not, av})
	require.Contains(t, strings.Join(av.GetChunkTags(), ","), "I-ADJP")
}
