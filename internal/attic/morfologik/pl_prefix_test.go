package morfologik

import (
	"path/filepath"
	"testing"
)

// Polish POS dict uses fsa.dict.encoder=PREFIX (TrimPrefixAndSuffixEncoder).
func TestPolishPREFIXLookup(t *testing.T) {
	p := filepath.Join(repoRoot(t), "inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl/polish.dict")
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	if d.Encoder != "PREFIX" {
		t.Fatalf("encoder=%s want PREFIX", d.Encoder)
	}
	cases := map[string]string{
		"był":     "być",
		"dam":     "dać",
		"ma":      "mieć",
		"wyszedł": "wyjść",
		"lepsze":  "dobry",
		"wywarł":  "wywrzeć",
		"darł":    "drzeć",
	}
	for w, wantLemma := range cases {
		forms, err := d.Lookup(w)
		if err != nil {
			t.Fatalf("%s: %v", w, err)
		}
		found := false
		for _, f := range forms {
			if f.Stem == wantLemma {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%q: want lemma %q in %+v", w, wantLemma, forms)
		}
	}
}
