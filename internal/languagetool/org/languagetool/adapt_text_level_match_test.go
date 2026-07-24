package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdaptTextLevelLocalMatch(t *testing.T) {
	sents := []SentenceData{
		NewSentenceData(nil, "Hello. ", 0, 0, 1),
		NewSentenceData(nil, "World", 7, 0, 8),
	}
	// match at offset 8 in "World" → second sentence
	m := LocalMatch{FromPos: 8, ToPos: 10, RuleID: "TL", Message: "x"}
	adj := AdaptTextLevelLocalMatch(m, sents, nil)
	require.Equal(t, 8, adj.FromPos)
	require.Equal(t, 10, adj.ToPos)
	require.Equal(t, 0, adj.Line)
	// column: FindLineColumn for offset 8 → prefix "W" in sent2, ProcessColumnChange(8,"W")=9
	// then line==0 → column - 1 = 8
	require.Equal(t, 8, adj.Column)
}

func TestAdaptTextLevelLocalMatch_WithMapper(t *testing.T) {
	sents := []SentenceData{NewSentenceData(nil, "abc", 0, 0, 1)}
	m := LocalMatch{FromPos: 0, ToPos: 2, RuleID: "TL"}
	// identity-like with +3 markup shift
	mapper := func(pos int, isToPos bool) int { return pos + 3 }
	adj := AdaptTextLevelLocalMatch(m, sents, mapper)
	require.Equal(t, 3, adj.FromPos)
	// toPos-1 mapped +1 → (2-1)+3+1 = 5
	require.Equal(t, 5, adj.ToPos)
}

func TestIgnoreRangesFromLanguageMatches(t *testing.T) {
	r, ok := IgnoreRangesFromLanguageMatches(10, 20, map[string]float32{"de": 0.9})
	require.True(t, ok)
	require.Equal(t, 10, r.FromPos)
	require.Equal(t, 20, r.ToPos)
	require.Equal(t, "de", r.Lang)

	_, ok = IgnoreRangesFromLanguageMatches(0, 1, nil)
	require.False(t, ok)

	ign := []Range{}
	ext := NewExtendedSentenceRange(10, 20, "en")
	ign = ApplyNewLanguageMatchesToSentence(ign, &ext, 10, 20, map[string]float32{"fr": 0.8})
	require.Len(t, ign, 1)
	require.Equal(t, "fr", ign[0].Lang)
	require.Equal(t, float32(0.8), ext.LanguageConfidenceRates["fr"])
	// unique
	ign = ApplyNewLanguageMatchesToSentence(ign, &ext, 10, 20, map[string]float32{"fr": 0.8})
	require.Len(t, ign, 1)
}
