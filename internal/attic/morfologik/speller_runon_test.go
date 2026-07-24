package morfologik

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestReplaceRunOnWordCandidates_EN(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	if !d.SupportRunOnWords {
		t.Fatal("en_US should support run-on (default true)")
	}
	// known word → empty
	if c := d.ReplaceRunOnWordCandidates("the"); len(c) != 0 {
		t.Fatalf("known word: %v", c)
	}
	// thecat → the cat
	words := d.ReplaceRunOnWords("thecat")
	if !slices.Contains(words, "the cat") {
		t.Fatalf("thecat → %v", words)
	}
	// weighted distance 1 class
	cds := d.ReplaceRunOnWordCandidates("thecat")
	found := false
	for _, c := range cds {
		if c.Word == "the cat" {
			found = true
			if c.OrigDistance != 1 {
				t.Fatalf("origDistance=%d want 1", c.OrigDistance)
			}
			// weight = 1*26 + 26 - freq - 1 = 51 - freq
			if c.Distance > 51 {
				t.Fatalf("distance weight=%d too high", c.Distance)
			}
		}
	}
	if !found {
		t.Fatalf("missing the cat in %v", cds)
	}
}

func TestReplaceRunOnWordCandidates_SentenceStart(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	// Thecat: prefix The via lower the
	words := d.ReplaceRunOnWords("Thecat")
	if !slices.Contains(words, "The cat") {
		t.Fatalf("Thecat → %v", words)
	}
}

func TestReplaceRunOnWordCandidates_Disabled(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	d.SupportRunOnWords = false
	if len(d.ReplaceRunOnWords("thecat")) != 0 {
		t.Fatal("disabled run-on")
	}
}
