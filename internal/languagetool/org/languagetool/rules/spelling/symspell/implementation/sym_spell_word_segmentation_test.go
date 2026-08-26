package implementation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SymSpell.wordSegmentation — Java goldens on fixture dict.
func TestWordSegmentation_JavaGoldens(t *testing.T) {
	s := fixtureDict(t)

	// seg helloworld: seg=hello world corr=hello world dist=0
	// plog=log10(11/N)+log10(5/N) ≈ -22.28100730310576
	got := s.WordSegmentation("helloworld")
	require.Equal(t, "hello world", got.SegmentedString)
	require.Equal(t, "hello world", got.CorrectedString)
	require.Equal(t, 0, got.DistanceSum)
	require.InDelta(t, -22.28100730310576, got.ProbabilityLogSum, 1e-12)

	// seg heloworld: seg=helo world corr=hello world dist=1
	got = s.WordSegmentation("heloworld")
	require.Equal(t, "helo world", got.SegmentedString)
	require.Equal(t, "hello world", got.CorrectedString)
	require.Equal(t, 1, got.DistanceSum)
	require.InDelta(t, -22.28100730310576, got.ProbabilityLogSum, 1e-12)

	// seg hello: exact
	got = s.WordSegmentation("hello")
	require.Equal(t, "hello", got.SegmentedString)
	require.Equal(t, "hello", got.CorrectedString)
	require.Equal(t, 0, got.DistanceSum)
	require.InDelta(t, -10.969292311141777, got.ProbabilityLogSum, 1e-12)
}

// Twin: empty input → empty result (Java new SegmentedSuggestion()).
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
	require.NotEmpty(t, got.CorrectedString)
	// Typo should not be corrected when maxEd=0
	got2 := s.WordSegmentationEdit("helo", 0)
	require.Equal(t, "helo", got2.CorrectedString)
}

// Twin: typo correction during segmentation.
func TestWordSegmentation_WithTypo(t *testing.T) {
	s := fixtureDict(t)
	got := s.WordSegmentation("heloworld")
	require.Equal(t, "hello world", got.CorrectedString)
	require.Equal(t, 1, got.DistanceSum)
}

// Twin: probability log uses N and is negative for rare words.
func TestWordSegmentation_ProbabilityLog(t *testing.T) {
	s := DefaultSymSpell()
	require.True(t, s.CreateDictionaryEntry("ab", 100, nil))
	got := s.WordSegmentation("ab")
	require.True(t, got.ProbabilityLogSum < 0, "log10(c/N) < 0: %v", got.ProbabilityLogSum)
	require.False(t, math.IsInf(got.ProbabilityLogSum, 0))
	require.False(t, math.IsNaN(got.ProbabilityLogSum))
	// exact: log10(100/N)
	want := math.Log10(100.0 / float64(corpusWordCountN))
	require.InDelta(t, want, got.ProbabilityLogSum, 1e-12)
}

// Twin: existing space in input is considered.
// Java (same fixture): correctedString/segmentedString = "hello  world" (double space
// from joining compositions with " " while the input space is also kept as a part), dist=0.
func TestWordSegmentation_ExistingSpace(t *testing.T) {
	s := fixtureDict(t)
	got := s.WordSegmentation("hello world")
	require.Equal(t, "hello  world", got.CorrectedString)
	require.Equal(t, "hello  world", got.SegmentedString)
	require.Equal(t, 0, got.DistanceSum)
}

// Twin: N constant matches Java private static long N.
func TestCorpusWordCountN(t *testing.T) {
	require.Equal(t, int64(1024908267229), corpusWordCountN)
}
