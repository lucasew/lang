package fa

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java rule demo sentences (addExamplePair) — correction = fixed marker span.
func TestFA_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"حاضر"}, NewSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"برای"}, NewPersianWordRepeatRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"این خیابان"}, NewPersianWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
