package morfologik

import (
	"path/filepath"
	"testing"
)

func TestGetFrequency_EnUS(t *testing.T) {
	root := freqRepoRoot(t)
	p := filepath.Join(root, "third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	if !d.FrequencyIncluded() {
		t.Fatal("expected frequency-included on en_US")
	}
	// known words should have non-negative frequency; common words often > 0
	fThe := d.GetFrequency("the")
	fHouse := d.GetFrequency("house")
	fXyz := d.GetFrequency("xyzzyqqq")
	t.Logf("the=%d house=%d xyz=%d", fThe, fHouse, fXyz)
	if fXyz != 0 {
		t.Fatalf("unknown want 0 got %d", fXyz)
	}
	// at least one common word should have frequency > 0 if encoding works
	if fThe == 0 && fHouse == 0 {
		// may still be 0 if last-byte encoding differs — probe forms
		forms, _ := d.Lookup("the")
		t.Fatalf("expected frequency for the/house; forms=%+v the=%d house=%d", forms, fThe, fHouse)
	}
}

func freqRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if st, err := filepath.Glob(filepath.Join(dir, "third_party", "english-pos-dict")); err == nil && len(st) > 0 {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
