package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of Java calcSpellerSuggestions: speller1 (edit-1) empty → speller2 (edit-2).
func TestBinaryCascadeSuggestions_Garentee(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	// Java: Speller(dict, maxEdit) alone does not cascade — rule uses speller1/2/3 Multis.
	sp1 := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp1.AttachBinaryDictionary(path))
	require.Empty(t, sp1.FindReplacements("garentee"), "edit-1 alone must not invent edit-2 hits")

	sp2 := NewMorfologikSpeller("/en/hunspell/en_US.dict", 2)
	require.True(t, sp2.AttachBinaryDictionary(path))
	sugs := sp2.FindReplacements("garentee")
	require.NotEmpty(t, sugs, "edit-2 speller should suggest for garentee")
	found := false
	for _, s := range sugs {
		if s == "guarantee" || s == "guaranteed" || s == "guarantees" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs)

	sugs2 := sp2.FindReplacements("greatful")
	require.NotEmpty(t, sugs2)
	found = false
	for _, s := range sugs2 {
		if s == "grateful" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs2)
}

// Twin of MorfologikSpellerRule.calcSpellerSuggestions cascade via Speller1/2/3 Multis.
func TestRuleCascadeSuggestions_Garentee(t *testing.T) {
	if DiscoverLanguageDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	s1 := OpenMultiSpellerFromClasspath("/en/hunspell/en_US.dict", nil, "", 1, nil)
	s2 := OpenMultiSpellerFromClasspath("/en/hunspell/en_US.dict", nil, "", 2, nil)
	s3 := OpenMultiSpellerFromClasspath("/en/hunspell/en_US.dict", nil, "", 3, nil)
	require.NotNil(t, s1)
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/hunspell/en_US.dict", nil)
	r.SetMultiSpellers(s1, s2, s3)

	sugs := r.collectSuggestions("garentee")
	require.NotEmpty(t, sugs, "rule cascade speller2 should suggest for garentee")
	found := false
	for _, s := range sugs {
		if s == "guarantee" || s == "guaranteed" || s == "guarantees" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs)

	sugs2 := r.collectSuggestions("greatful")
	require.NotEmpty(t, sugs2)
	found = false
	for _, s := range sugs2 {
		if s == "grateful" {
			found = true
		}
	}
	require.True(t, found, "sugs=%v", sugs2)

	// speller1-only still finds edit-1
	sugs3 := r.collectSuggestions("recieve")
	require.Contains(t, sugs3, "receive")
}
