package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of Java calcSpellerSuggestions: edit-2 cascade when edit-1 empty.
func TestBinaryCascadeSuggestions_Garentee(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	sugs := sp.FindReplacements("garentee")
	require.NotEmpty(t, sugs, "edit-2 cascade should suggest for garentee")
	// prefer guarantee among results
	found := false
	for _, s := range sugs {
		if s == "guarantee" || s == "guaranteed" || s == "guarantees" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs)

	sugs2 := sp.FindReplacements("greatful")
	require.NotEmpty(t, sugs2)
	found = false
	for _, s := range sugs2 {
		if s == "grateful" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs2)
}
