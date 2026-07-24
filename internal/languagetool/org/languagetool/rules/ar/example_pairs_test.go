package ar

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestAR_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"لن"}, NewArabicRedundancyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"تجرِِبة"}, NewArabicDiacriticsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"فلفل حلو"}, NewArabicDarjaRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"في العبارة خطأ"}, NewArabicWordinessRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"ظن"}, NewArabicHomophonesRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"إلى"}, NewArabicSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"شؤون"}, NewArabicWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"فقط"}, NewArabicWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"الظن"}, NewArabicWrongWordInContextRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"بحوثا"}, NewArabicInflectedOneWordReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"an hour"}, NewArabicTransVerbRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"الظن"}, NewArabicConfusionProbabilityRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
