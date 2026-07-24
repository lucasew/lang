package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeSentenceData(t *testing.T) {
	// Two sentences: "Hello. " (7) then "World" (5)
	a0 := NewAnalyzedSentence([]*AnalyzedTokenReadings{
		NewAnalyzedTokenReadings(NewAnalyzedToken("Hello", nil, nil)),
	})
	a1 := NewAnalyzedSentence([]*AnalyzedTokenReadings{
		NewAnalyzedTokenReadings(NewAnalyzedToken("World", nil, nil)),
	})
	sents := ComputeSentenceData([]*AnalyzedSentence{a0, a1}, []string{"Hello. ", "World"}, false)
	require.Len(t, sents, 2)

	require.Equal(t, 0, sents[0].StartOffset)
	require.Equal(t, 0, sents[0].StartLine)
	require.Equal(t, 1, sents[0].StartColumn) // Java columnCount starts at 1

	require.Equal(t, 7, sents[1].StartOffset) // UTF-16 len("Hello. ")
	require.Equal(t, 0, sents[1].StartLine)
	// processColumnChange(1, "Hello. ") → no nl → 1+7 = 8
	require.Equal(t, 8, sents[1].StartColumn)

	// With newline: second sentence after line break
	sents2 := ComputeSentenceData(nil, []string{"Hi\n", "There"}, false)
	require.Equal(t, 0, sents2[0].StartOffset)
	require.Equal(t, 0, sents2[0].StartLine)
	require.Equal(t, 1, sents2[0].StartColumn)
	require.Equal(t, 3, sents2[1].StartOffset) // "Hi\n"
	require.Equal(t, 1, sents2[1].StartLine)
	// processColumnChange: last nl at 2, len 3 → column = 3-2 = 1; lineBreakPos!=0 so no --
	require.Equal(t, 1, sents2[1].StartColumn)
}

func TestComputeSentenceData_Empty(t *testing.T) {
	require.Nil(t, ComputeSentenceData(nil, nil, false))
	require.Nil(t, ComputeSentenceData([]*AnalyzedSentence{}, []string{}, false))
}

func TestBuildExtendedSentenceRange(t *testing.T) {
	// Normal sentence, no double leading space
	sd := NewSentenceData(nil, "Hello world", 10, 0, 1)
	r := BuildExtendedSentenceRange(sd, "en")
	require.Equal(t, 10, r.FromPos)
	require.Equal(t, 10+11, r.ToPos) // trim no-op
	require.Equal(t, float32(1.0), r.LanguageConfidenceRates["en"])

	// startsWith(" ", 1): second char is space → "  foo"
	sd2 := NewSentenceData(nil, "  foo", 0, 0, 1)
	r2 := BuildExtendedSentenceRange(sd2, "de")
	// stripLeading → "foo", whitespaceFix = 5-3 = 2
	require.Equal(t, 2, r2.FromPos)
	// trim("  foo") = "foo" len 3 → to = 2+3 = 5
	require.Equal(t, 5, r2.ToPos)

	// single leading space: startsWith(" ", 1) is false (charAt(1)='H' for " Hello")
	sd3 := NewSentenceData(nil, " Hello", 0, 0, 1)
	r3 := BuildExtendedSentenceRange(sd3, "en")
	require.Equal(t, 0, r3.FromPos) // no whitespaceFix
	// trim removes leading space → "Hello" len 5
	require.Equal(t, 5, r3.ToPos)
}

func TestExtendedSentenceRange_HashCode(t *testing.T) {
	a := NewExtendedSentenceRange(1, 5, "en")
	b := NewExtendedSentenceRange(1, 5, "de")
	require.True(t, a.Equal(b))
	require.Equal(t, a.HashCode(), b.HashCode())
	require.Equal(t, 31*1+5, a.HashCode())
}
