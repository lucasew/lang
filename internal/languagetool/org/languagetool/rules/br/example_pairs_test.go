package br

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBR_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"alc'hweder-gwez"}, NewBretonCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
