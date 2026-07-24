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

	// Full Java map smokes
	require.Contains(t, FindTagsAdj("5", "х"), ":p:v_rod")
	require.Contains(t, FindTagsAdj("5", "ій"), ":f:v_dav") // not 3 → f only
	require.Contains(t, FindTagsAdj("3", "ій"), ":m:v_naz") // 3 → always branch
	require.Contains(t, FindTagsNoun("20", "ти"), ":p:v_rod:bad")
	require.Empty(t, FindTagsNoun("1", "ти")) // pattern requires [0569]|1[0-9]
	require.True(t, IsPossibleAdjAdjEnding("1", "й"))
}
