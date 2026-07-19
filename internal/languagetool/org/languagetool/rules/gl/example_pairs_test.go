package gl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestGL_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"currículo"}, NewGalicianBarbarismsRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"duna"}, NewGalicianRedundancyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"para os efectos de"}, NewGalicianWikipediaRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Raramente acontece"}, NewGalicianWordinessRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
