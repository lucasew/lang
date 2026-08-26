package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestES_Wikipedia_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"abasto"}, NewSpanishWikipediaRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
