package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckResults(t *testing.T) {
	ext := []ExtendedSentenceRange{
		NewExtendedSentenceRange(10, 20, "en"),
		NewExtendedSentenceRange(0, 5, "en"),
	}
	cr := NewCheckResultsFull(nil, []Range{NewRange(1, 2, "en")}, ext)
	// sorted by fromPos
	require.Equal(t, 0, cr.GetExtendedSentenceRanges()[0].FromPos)
	require.Equal(t, 10, cr.GetExtendedSentenceRanges()[1].FromPos)
	cr.AddSentenceRanges([]SentenceRange{NewSentenceRange(0, 5)})
	require.Len(t, cr.GetSentenceRanges(), 1)
}
