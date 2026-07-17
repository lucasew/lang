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
	sugs := d.SuggestEdits("Recieve", 8)
	if !slices.Contains(sugs, "Receive") {
		t.Fatalf("expected Receive among %v", sugs)
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
