package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.ExtendedSentenceRange (no upstream unit test; from source).

func TestExtendedSentenceRange(t *testing.T) {
	a := NewExtendedSentenceRange(10, 20, "en")
	require.Equal(t, 10, a.GetFromPos())
	require.Equal(t, 20, a.GetToPos())
	require.Equal(t, float32(1.0), a.GetLanguageConfidenceRates()["en"])

	// equals: only fromPos/toPos (not language map)
	b := NewExtendedSentenceRangeWithRates(10, 20, map[string]float32{"de": 0.5})
	require.True(t, a.Equal(b))
	require.False(t, a.Equal(NewExtendedSentenceRange(11, 20, "en")))

	// compareTo / Less: by fromPos
	require.True(t, NewExtendedSentenceRange(0, 5, "en").Less(a))
	require.False(t, a.Less(NewExtendedSentenceRange(0, 5, "en")))

	// updateLanguageConfidenceRates replaces map
	a.UpdateLanguageConfidenceRates(map[string]float32{"fr": 0.9})
	require.Equal(t, float32(0.9), a.GetLanguageConfidenceRates()["fr"])
	_, hasEn := a.GetLanguageConfidenceRates()["en"]
	require.False(t, hasEn)

	require.Contains(t, a.String(), "10-20:")
}
