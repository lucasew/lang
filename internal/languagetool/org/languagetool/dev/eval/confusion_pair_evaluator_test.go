package eval

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfusionPairEvaluator_TP_FP(t *testing.T) {
	// word0=there, word1=their; rule0=R_THERE, rule1=R_THEIR
	// When correct has "there", rule R_THERE should not fire (TN); swapped "their" should fire R_THEIR (TP)
	ev := NewConfusionPairEvaluator("there", "their", "R_THERE", "R_THEIR",
		func(sentence string) ([]string, error) {
			var ids []string
			if strings.Contains(sentence, "their") {
				ids = append(ids, "R_THEIR") // detects wrong their when should be there?
			}
			// simplistic: fire R_THEIR whenever their present; R_THERE whenever there present as false positive sometimes
			if strings.Contains(sentence, "there") && !strings.Contains(sentence, "their") {
				// correct "there" — no match (TN for R_THERE)
			}
			return ids, nil
		})
	require.NoError(t, ev.ProcessLine("I go there often."))
	// j=0: correct with there → TN for R_THERE; wrong with their → TP for R_THEIR
	require.Equal(t, 1, ev.Results[0][classTN])
	require.Equal(t, 1, ev.Results[1][classTP])
	require.InDelta(t, 1.0, ev.Precision(1), 1e-9)
	require.InDelta(t, 1.0, ev.Recall(1), 1e-9)
}

func TestConfusionPairEvaluator_FP(t *testing.T) {
	// rule wrongly fires on correct sentence
	ev := NewConfusionPairEvaluator("a", "an", "R_A", "R_AN",
		func(sentence string) ([]string, error) {
			if strings.Contains(sentence, " a ") || strings.HasPrefix(sentence, "a ") {
				return []string{"R_A"}, nil // always flag a → FP on correct
			}
			if strings.Contains(sentence, " an ") || strings.HasPrefix(sentence, "an ") {
				return []string{"R_AN"}, nil
			}
			return nil, nil
		})
	require.NoError(t, ev.ProcessLine("I saw a cat."))
	require.Equal(t, 1, ev.Results[0][classFP])
}
