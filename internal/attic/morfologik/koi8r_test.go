package morfologik

import (
	"path/filepath"
	"testing"
)

func findRussianDict(t *testing.T) string {
	t.Helper()
	// walk from cwd
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
		"src", "main", "resources", "org", "languagetool", "resource", "ru", "russian.dict")
	// try a few parents
	p := rel
	for i := 0; i < 6; i++ {
		if st, err := filepath.Abs(p); err == nil {
			if _, e := filepath.Glob(st); e == nil {
				if _, err := OpenDictionary(st); err == nil {
					// check exists
					if d, err := OpenDictionary(st); err == nil && d != nil {
						return st
					}
				}
			}
		}
		// simpler: just Stat
		if _, err := filepath.Abs(p); err == nil {
			// use Open only
		}
		cand := p
		if d, err := OpenDictionary(cand); err == nil && d != nil {
			return cand
		}
		p = filepath.Join("..", p)
	}
	// absolute from known workspace layout relative to module
	candidates := []string{
		rel,
		filepath.Join("..", rel),
		filepath.Join("../..", rel),
		filepath.Join("../../..", rel),
		filepath.Join("../../../..", rel),
	}
	for _, c := range candidates {
		if d, err := OpenDictionary(c); err == nil && d != nil {
			return c
		}
	}
	t.Skip("russian.dict not found")
	return ""
}

func TestRussianKOI8RLookup(t *testing.T) {
	p := findRussianDict(t)
	d, err := OpenDictionary(p)
	if err != nil {
		t.Fatal(err)
	}
	if d.Encoding != "koi8-r" {
		t.Fatalf("encoding=%q want koi8-r", d.Encoding)
	}
	forms, err := d.Lookup("дом")
	if err != nil {
		t.Fatal(err)
	}
	if len(forms) == 0 {
		t.Fatal("expected POS forms for дом")
	}
	t.Logf("дом -> %+v", forms)
}
