package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLetterEndingForNumericHelper(t *testing.T) {
	var h LetterEndingForNumericHelper
	require.True(t, h.HasKnownEnding("й"))
	tags := h.TagsForAdjEnding("й", "1")
	require.Contains(t, tags, ":m:v_naz")
	require.Empty(t, h.TagsForAdjEnding("zzz", "1"))
}
