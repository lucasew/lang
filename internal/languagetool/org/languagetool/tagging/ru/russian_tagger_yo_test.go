package ru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRussianTagger_MayMissingYO(t *testing.T) {
	if DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	EnsureDefaultRussianTagger()
	// "все" can be missing ё (всё): flag only if е→ё form is in dict
	got := DefaultRussianTagger.Tag([]string{"все"})
	require.Len(t, got, 1)
	// lowercased е→ё probe: всё should exist in dict → MayMissingYO
	require.Contains(t, got[0].GetChunkTags(), "MayMissingYO")
	// word with ё already → no flag
	got2 := DefaultRussianTagger.Tag([]string{"всё"})
	require.NotContains(t, got2[0].GetChunkTags(), "MayMissingYO")
	// no е → no flag
	got3 := DefaultRussianTagger.Tag([]string{"дом"})
	require.NotContains(t, got3[0].GetChunkTags(), "MayMissingYO")
}
