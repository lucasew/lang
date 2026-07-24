package morfologik

import (
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	// internal/attic/morfologik -> repo root is 3 up
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func TestFSA5DanishLookup(t *testing.T) {
	p := filepath.Join(repoRoot(t), "inspiration/languagetool/languagetool-language-modules/da/src/main/resources/org/languagetool/resource/da/danish.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	if d.FSA.version != versionFSA5 {
		t.Fatalf("version 0x%02x want FSA5", d.FSA.version)
	}
	forms, err := d.Lookup("ham")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ham => %+v", forms)
	if len(forms) == 0 {
		t.Fatal("expected tags for ham")
	}
	foundPron := false
	for _, f := range forms {
		t.Logf("  stem=%q tag=%q", f.Stem, f.Tag)
		if len(f.Tag) >= 4 && (f.Tag[:4] == "pron" || f.Tag[:3] == "pro") {
			foundPron = true
		}
	}
	if !foundPron {
		t.Fatal("expected pron tag on ham")
	}
}

func TestFSA5KhmerLookup(t *testing.T) {
	p := filepath.Join(repoRoot(t), "inspiration/languagetool/languagetool-language-modules/km/src/main/resources/org/languagetool/resource/km/khmer.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	forms, err := d.Lookup("ខ្ញុំ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ខ្ញុំ => %+v", forms)
	if len(forms) == 0 {
		t.Fatal("expected tags")
	}
}

func TestCFSA2EnglishStillWorks(t *testing.T) {
	p := filepath.Join(repoRoot(t), "third_party/english-pos-dict/org/languagetool/resource/en/english.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Skip(err)
	}
	forms, err := d.Lookup("the")
	if err != nil || len(forms) == 0 {
		t.Fatalf("the: %v %v", forms, err)
	}
	t.Logf("the => %+v", forms)
}
