package morfologik

import (
	"os"
	"path/filepath"
	"testing"
)

func findDict(t *testing.T, rel string) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for {
		for _, base := range []string{
			filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource"),
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource"),
		} {
			p := filepath.Join(base, rel)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("dict not found:", rel)
		}
		dir = parent
	}
}

func TestEnglishPOSLookup(t *testing.T) {
	p := findDict(t, "en/english.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	forms, err := d.Lookup("houses")
	if err != nil {
		t.Fatal(err)
	}
	if len(forms) == 0 {
		t.Fatal("expected analyses for 'houses'")
	}
	t.Logf("houses -> %+v", forms)
	foundNN := false
	for _, f := range forms {
		if f.Tag != "" {
			foundNN = true
		}
	}
	if !foundNN {
		t.Fatalf("expected POS tags, got %+v", forms)
	}
}

func TestEnglishSpellerContains(t *testing.T) {
	p := findDict(t, "en/hunspell/en_US.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	// Speller dicts use word + separator + encoded frequency/info
	forms, err := d.Lookup("house")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("house forms=%+v contains=%v", forms, d.Contains("house"))
	if !d.Contains("house") && len(forms) == 0 {
		// try lowercase common words
		for _, w := range []string{"the", "house", "computer", "language"} {
			f, _ := d.Lookup(w)
			t.Logf("%q lookup=%+v exact=%v", w, f, func() bool {
				k, _, _ := d.FSA.Match([]byte(w), d.FSA.RootNode())
				return k == ExactMatch || k == SequenceIsAPrefix
			}())
		}
		t.Fatal("speller did not recognize 'house'")
	}
}
