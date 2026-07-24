package nl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNL_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"organogram"}, NewWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"klaar"}, NewSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"fiets"}, NewPreferredWordRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
