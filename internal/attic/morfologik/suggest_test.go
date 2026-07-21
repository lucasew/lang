package morfologik

import (
	"slices"
	"testing"
)

func TestSuggestEdits_Recieve(t *testing.T) {
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	sugs := d.SuggestEdits("recieve", 8)
	if !slices.Contains(sugs, "receive") {
		t.Fatalf("expected receive among %v", sugs)
	}
}

func TestSuggestEdits_TitleCase(t *testing.T) {
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	// Java Speller.findReplacements returns dictionary forms (lowercase); case fold is
	// MorfologikSpeller.getSuggestions, not Speller.
	sugs := d.SuggestEdits("Recieve", 8)
	if !slices.Contains(sugs, "receive") {
		t.Fatalf("expected receive among %v", sugs)
	}
}

func TestSuggestEdits_ShortReplacement_Fone(t *testing.T) {
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.ReplacementShort) == 0 {
		t.Fatal("expected replacement-pairs loaded from en_US.info")
	}
	sugs := d.SuggestEdits("fone", 8)
	if !slices.Contains(sugs, "phone") {
		t.Fatalf("fone → phone via anyToTwo; got %v", sugs)
	}
}

func TestSuggestEdits_Empty(t *testing.T) {
	var nilD *Dictionary
	if nilD.SuggestEdits("x", 8) != nil {
		t.Fatal("nil dict")
	}
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	if d.SuggestEdits("", 8) != nil {
		t.Fatal("empty word")
	}
}
