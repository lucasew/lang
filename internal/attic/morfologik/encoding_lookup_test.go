package morfologik

import (
	"path/filepath"
	"testing"
)

func openWalkDict(t *testing.T, rel string) *Dictionary {
	t.Helper()
	candidates := []string{rel}
	for i := 0; i < 6; i++ {
		candidates = append(candidates, filepath.Join("..", candidates[len(candidates)-1]))
	}
	// also from repo roots
	candidates = append(candidates,
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", rel),
	)
	// rebuild candidates properly
	parts := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", rel),
	}
	for i := 0; i < 5; i++ {
		parts = append(parts, filepath.Join("..", parts[len(parts)-1]))
	}
	for _, c := range parts {
		if d, err := OpenDictionary(c); err == nil && d != nil {
			return d
		}
	}
	t.Skipf("dict not found: %s", rel)
	return nil
}

func TestItalianISO8859_15Lookup(t *testing.T) {
	d := openWalkDict(t, filepath.Join("it", "src", "main", "resources", "org", "languagetool", "resource", "it", "italian.dict"))
	if got := d.Encoding; got != "iso-8859-15" && got != "ISO-8859-15" {
		t.Fatalf("encoding=%q", got)
	}
	// common Italian words
	for _, w := range []string{"casa", "sono", "il", "della"} {
		forms, err := d.Lookup(w)
		if err != nil {
			t.Fatal(err)
		}
		if len(forms) == 0 {
			t.Fatalf("no forms for %q", w)
		}
		t.Logf("%s -> %+v", w, forms[:1])
	}
}

func TestTagalogISO8859_1Lookup(t *testing.T) {
	d := openWalkDict(t, filepath.Join("tl", "src", "main", "resources", "org", "languagetool", "resource", "tl", "tagalog.dict"))
	forms, err := d.Lookup("ako") // common Tagalog word "I"
	if err != nil {
		t.Fatal(err)
	}
	// may or may not be in dict; just ensure no panic and encoding applied
	t.Logf("encoding=%s forms=%d", d.Encoding, len(forms))
	if d.charset() == nil {
		t.Fatal("iso-8859-1 should resolve charset")
	}
}
