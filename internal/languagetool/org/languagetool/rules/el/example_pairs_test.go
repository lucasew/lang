package el

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestEL_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"Ηνωμένες Πολιτείες"}, NewGreekSpecificCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"ανεβαίνω"}, NewGreekRedundancyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Επιπλέον"}, NewGreekWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"20ός"}, NewNumeralStressRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
