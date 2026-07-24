package eval

// Twin of RealWordCorpusEvaluatorTest — green inject scoring
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/errorcorpus"
	"github.com/stretchr/testify/require"
)

// Port of RealWordCorpusEvaluatorTest.testCheck (Java @Ignore lifted with inject)
func TestRealWordCorpusEvaluator_Check(t *testing.T) {
	// gold: "teh" → "the" at 0:3 in "teh cat"
	ev := NewRealWordCorpusEvaluator(FuncEvaluator{Fn: func(text string) ([]Match, error) {
		return []Match{{
			FromPos:               0,
			ToPos:                 3,
			SuggestedReplacements: []string{"the"},
		}}, nil
	}})
	require.NoError(t, ev.CheckSentence(CorpusSentence{
		PlainText: "teh cat",
		Errors:    []GoldError{{StartPos: 0, EndPos: 3, Correction: "the"}},
	}))
	require.Equal(t, 1, ev.GetSentencesChecked())
	require.Equal(t, 1, ev.GetErrorsChecked())
	require.Equal(t, 1, ev.MatchCount)
	require.Equal(t, 1, ev.GetRealErrorsFound())
	require.Equal(t, 1, ev.GetRealErrorsFoundWithGoodSuggestion())
	pr := ev.AnySuggestionPR()
	require.InDelta(t, 1.0, pr.GetPrecision(), 1e-9)
	require.InDelta(t, 1.0, pr.GetRecall(), 1e-9)
	require.InDelta(t, 1.0, pr.F05(), 1e-9)

	// false positive match
	ev2 := NewRealWordCorpusEvaluator(FuncEvaluator{Fn: func(text string) ([]Match, error) {
		return []Match{{FromPos: 4, ToPos: 7, SuggestedReplacements: []string{"dog"}}}, nil
	}})
	require.NoError(t, ev2.CheckSentence(CorpusSentence{
		PlainText: "teh cat",
		Errors:    []GoldError{{StartPos: 0, EndPos: 3, Correction: "the"}},
	}))
	require.Equal(t, 0, ev2.GoodMatches)
	require.Equal(t, 1, ev2.MatchCount)
	require.Equal(t, 0.0, ev2.PrecisionAnySuggestion())
}

func TestFromErrorSentence(t *testing.T) {
	es := errorcorpus.NewErrorSentence("teh cat", []errorcorpus.Error{
		{StartPos: 0, EndPos: 3, Correction: "the"},
	})
	es.PlainText = "teh cat"
	cs := FromErrorSentence(es)
	require.Equal(t, "teh cat", cs.PlainText)
	require.Len(t, cs.Errors, 1)
	require.Equal(t, "the", cs.Errors[0].Correction)
}
