package implementation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SymSpell.wordSegmentation — insert space between known words.
func TestWordSegmentation_JoinKnown(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 1000000, nil))
	require.True(t, s.CreateDictionaryEntry("world", 800000, nil))
	// maxLength after entries is len of longest term
	got := s.WordSegmentation("helloworld")
	require.NotEmpty(t, got.CorrectedString)
	// Expect space-separated known words (order preserved)
	require.Contains(t, got.CorrectedString, "hello")
	require.Contains(t, got.CorrectedString, "world")
	require.Contains(t, got.SegmentedString, "hello")
}

// Twin: empty input → empty result.
func TestWordSegmentation_Empty(t *testing.T) {
	s := DefaultSymSpell()
	got := s.WordSegmentation("")
	require.Equal(t, SegmentedSuggestion{}, got)
}

// Twin: single known word unchanged.
func TestWordSegmentation_SingleWord(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("test", 50, nil))
	got := s.WordSegmentation("test")
	require.Equal(t, "test", got.CorrectedString)
	require.Equal(t, "test", got.SegmentedString)
	require.Equal(t, 0, got.DistanceSum)
}

// Twin: maxEditDistance 0 — segmentation only, no correction.
func TestWordSegmentation_NoCorrection(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 100, nil))
	require.True(t, s.CreateDictionaryEntry("world", 80, nil))
	got := s.WordSegmentationEdit("helloworld", 0)
	// With ed=0, only exact dict matches segment
	require.NotEmpty(t, got.CorrectedString)
	// Typo should not be corrected when maxEd=0
	got2 := s.WordSegmentationEdit("helo", 0)
	// may stay "helo" as unknown word path
	require.Equal(t, "helo", got2.CorrectedString)
}

// Twin: typo correction during segmentation.
func TestWordSegmentation_WithTypo(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 1000000, nil))
	require.True(t, s.CreateDictionaryEntry("world", 800000, nil))
	got := s.WordSegmentation("heloworld")
	require.Contains(t, got.CorrectedString, "hello")
	require.Contains(t, got.CorrectedString, "world")
}

// Twin: probability log uses N and is negative for rare words.
func TestWordSegmentation_ProbabilityLog(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("ab", 100, nil))
	got := s.WordSegmentation("ab")
	require.True(t, got.ProbabilityLogSum < 0, "log10(c/N) < 0: %v", got.ProbabilityLogSum)
	// finite
	require.False(t, math.IsInf(got.ProbabilityLogSum, 0))
	require.False(t, math.IsNaN(got.ProbabilityLogSum))
}

// Twin: existing space in input is considered.
func TestWordSegmentation_ExistingSpace(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("hello", 100, nil))
	require.True(t, s.CreateDictionaryEntry("world", 80, nil))
	got := s.WordSegmentation("hello world")
	require.Contains(t, got.CorrectedString, "hello")
	require.Contains(t, got.CorrectedString, "world")
}
