package morfologik

import (
	"testing"
)

// TestSpeller_ContainsSeparatorsSticky ports Java Speller.containsSeparators mutation.
// Plain FSABuilder words are ExactMatch → first isInDictionary clears the flag permanently.
func TestSpeller_ContainsSeparatorsSticky_PlainFSA(t *testing.T) {
	d := NewDictionaryFromWords([]string{"receive", "recipe", "the", "cat"}, nil)
	if d == nil {
		t.Fatal("nil dict")
	}
	sp := NewSpeller(d, 1)
	if !sp.ContainsSeparators() {
		t.Fatal("default containsSeparators true")
	}
	if !sp.IsInDictionary("receive") {
		t.Fatal("receive must be in plain FSA")
	}
	// Java: ExactMatch without separator char → containsSeparators = false forever
	if sp.ContainsSeparators() {
		t.Fatal("after ExactMatch word, containsSeparators must be false (Java sticky)")
	}
	// Still accepts ExactMatch words
	if !sp.IsInDictionary("recipe") {
		t.Fatal("recipe still in dict after sticky false")
	}
	// Prefix of a word is SEQUENCE_IS_A_PREFIX; with separators=false must not use sep path
	if sp.IsInDictionary("receiv") {
		t.Fatal("prefix of receive must not be in-dict without separator arcs")
	}
	// Suggest still works with sticky false (isArcFinal ends candidates)
	cds := sp.FindReplacementCandidatesFull("recieve", false)
	found := false
	for _, c := range cds {
		if c.Word == "receive" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected receive in suggestions for recieve, got %+v", cds)
	}
}

// TestSpeller_ContainsSeparatorsSticky_BinaryEN: frequency-encoded en_US stays true.
func TestSpeller_ContainsSeparatorsSticky_BinaryEN(t *testing.T) {
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	sp := NewSpeller(d, 1)
	if !sp.IsInDictionary("house") {
		t.Fatal("house in en_US")
	}
	// SEQUENCE_IS_A_PREFIX + sep arc → never ExactMatch on normal words → flag stays true
	if !sp.ContainsSeparators() {
		t.Fatal("en_US speller must keep containsSeparators true (freq after separator)")
	}
	if sp.IsMisspelled("house") {
		t.Fatal("house not misspelled")
	}
	if !sp.IsMisspelled("xyzzyqqqnotaword") {
		t.Fatal("unknown word misspelled")
	}
}

// TestDictionary_ContainsCold_NoSharedMutation: cached Dictionary stays immutable.
func TestDictionary_ContainsCold_NoSharedMutation(t *testing.T) {
	d := NewDictionaryFromWords([]string{"receive", "recipe"}, nil)
	if !d.Contains("receive") {
		t.Fatal("cold contains")
	}
	// Second cold call still works (no sticky state on Dictionary)
	if !d.Contains("recipe") {
		t.Fatal("cold contains recipe")
	}
	sp := NewSpeller(d, 1)
	_ = sp.IsInDictionary("receive")
	if sp.ContainsSeparators() {
		t.Fatal("speller sticky false")
	}
	// Dictionary cold path independent of Speller instance
	if !d.Contains("receive") {
		t.Fatal("dictionary cold still works after Speller mutation")
	}
}
