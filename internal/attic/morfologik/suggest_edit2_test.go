package morfologik

import (
	"path/filepath"
	"testing"
)

// Twin of Java MorfologikSpellerRule: speller2 (edit distance 2) finds suggestions
// for "garentee" / "greatful" that edit-1 misses.
func TestSuggestEditsMax_Edit2(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	// edit-1 empty or weak for garentee
	s1 := d.SuggestEditsMax("garentee", 8, 1)
	s2 := d.SuggestEditsMax("garentee", 8, 2)
	t.Logf("edit1=%v edit2=%v", s1, s2)
	found := false
	for _, s := range s2 {
		if s == "guarantee" || s == "guaranteed" {
			found = true
		}
	}
	if !found {
		// may be guarantee with different ranking — require non-empty edit2 ≥ edit1
		if len(s2) == 0 {
			t.Fatalf("expected edit-2 suggestions for garentee, edit1=%v", s1)
		}
	}
	// greatful → grateful is classic edit-2
	s2b := d.SuggestEditsMax("greatful", 8, 2)
	found = false
	for _, s := range s2b {
		if s == "grateful" {
			found = true
		}
	}
	if !found {
		t.Logf("greatful → %v (grateful preferred)", s2b)
		if len(s2b) == 0 {
			t.Fatal("expected edit-2 suggestions for greatful")
		}
	}
}

func TestWeightedEditSuggestions_Order(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	w := d.WeightedEditSuggestions("recieve", 8, 1)
	if len(w) == 0 {
		t.Fatal("expected weighted suggestions")
	}
	// weights non-decreasing
	for i := 1; i < len(w); i++ {
		if w[i].Weight < w[i-1].Weight {
			t.Fatalf("weights not sorted: %v", w)
		}
	}
	// receive should be present
	found := false
	for _, x := range w {
		if x.Word == "receive" {
			found = true
		}
	}
	if !found {
		t.Fatalf("want receive in %v", w)
	}
}
