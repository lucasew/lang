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

func TestEnglishChunker_KeepSeeAndIsGoing(t *testing.T) {
	prp, vb, vbp, nn, vbz, vbg := "PRP", "VB", "VBP", "NN", "VBZ", "VBG"
	i := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("I", &prp, nil))
	keep := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("keep", &nn, nil),
		languagetool.NewAnalyzedToken("keep", &vb, nil),
		languagetool.NewAnalyzedToken("keep", &vbp, nil),
	}, 0)
	see := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("see", &nn, nil),
		languagetool.NewAnalyzedToken("see", &vb, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{i, keep, see})
	require.Contains(t, strings.Join(see.GetChunkTags(), ","), "VP")

	is := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("is", &vbz, nil))
	going := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("going", &nn, nil),
		languagetool.NewAnalyzedToken("going", &vbg, nil),
	}, 0)
	makeT := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("make", &nn, nil),
		languagetool.NewAnalyzedToken("make", &vb, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{is, going, makeT})
	require.Contains(t, strings.Join(going.GetChunkTags(), ","), "VP")
}

// OpenNLP (Java EnglishChunker): Let's hang out → hang B-VP, out B-PRT.
// LT dict pre-disambig tags 's as POS|VBZ only; soft primaryPOS must force PRP
// after "let" so PHRASAL_VERB_SOMETIME (chunk_re=".-VP" + chunk="B-PRT") fires.
func TestEnglishChunker_LetsHangOut(t *testing.T) {
	nn, vb, vbp, vbz, pos, rp, in, rb, uh := "NN", "VB", "VBP", "VBZ", "POS", "RP", "IN", "RB", "UH"
	let := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Let", &nn, nil),
		languagetool.NewAnalyzedToken("Let", &vb, nil),
		languagetool.NewAnalyzedToken("Let", &vbp, nil),
	}, 0)
	// Pre-disambiguation readings (no PRP yet — matches English.dict).
	s := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("'s", &pos, nil),
		languagetool.NewAnalyzedToken("'s", &vbz, nil),
	}, 0)
	hang := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("hang", &nn, nil),
		languagetool.NewAnalyzedToken("hang", &vb, nil),
		languagetool.NewAnalyzedToken("hang", &vbp, nil),
	}, 0)
	out := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("out", &in, nil),
		languagetool.NewAnalyzedToken("out", &nn, nil),
		languagetool.NewAnalyzedToken("out", &rb, nil),
		languagetool.NewAnalyzedToken("out", &rp, nil),
		languagetool.NewAnalyzedToken("out", &uh, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{let, s, hang, out})
	require.Contains(t, strings.Join(hang.GetChunkTags(), ","), "VP")
	require.Contains(t, strings.Join(out.GetChunkTags(), ","), "B-PRT")
}

// "and catch up" — particle lookahead forces catch VP + up B-PRT.
func TestEnglishChunker_AndCatchUp(t *testing.T) {
	nn, vb, vbp, cc, rp, in := "NN", "VB", "VBP", "CC", "RP", "IN"
	and := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("and", &cc, nil))
	catch := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("catch", &nn, nil),
		languagetool.NewAnalyzedToken("catch", &vb, nil),
		languagetool.NewAnalyzedToken("catch", &vbp, nil),
	}, 0)
	up := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("up", &in, nil),
		languagetool.NewAnalyzedToken("up", &rp, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{and, catch, up})
	require.Contains(t, strings.Join(catch.GetChunkTags(), ","), "VP")
	require.Contains(t, strings.Join(up.GetChunkTags(), ","), "B-PRT")
}

// "find out where" — particle before WRB stays I-VP for WHERE_MD_VB.
func TestEnglishChunker_FindOutWhere(t *testing.T) {
	vb, rp, in, wrb := "VB", "RP", "IN", "WRB"
	find := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("find", &vb, nil))
	out := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("out", &in, nil),
		languagetool.NewAnalyzedToken("out", &rp, nil),
	}, 0)
	where := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("where", &wrb, nil))
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{find, out, where})
	require.Contains(t, strings.Join(out.GetChunkTags(), ","), "VP")
}

// "door behind" — behind is prep so door is E-NP (LOOK_DOOR).
func TestEnglishChunker_DoorBehind(t *testing.T) {
	dt, nn, in, jj, rb, rp := "DT", "NN", "IN", "JJ", "RB", "RP"
	the := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("the", &dt, nil))
	door := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("door", &nn, nil))
	behind := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("behind", &in, nil),
		languagetool.NewAnalyzedToken("behind", &jj, nil),
		languagetool.NewAnalyzedToken("behind", &nn, nil),
		languagetool.NewAnalyzedToken("behind", &rb, nil),
		languagetool.NewAnalyzedToken("behind", &rp, nil),
	}, 0)
	NewEnglishChunker().AddChunkTags([]*languagetool.AnalyzedTokenReadings{the, door, behind})
	require.Contains(t, strings.Join(door.GetChunkTags(), ","), "E-NP")
	require.Contains(t, strings.Join(behind.GetChunkTags(), ","), "PP")
}
