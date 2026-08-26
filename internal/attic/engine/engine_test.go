package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func dataDir(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for {
		p := filepath.Join(dir, "inspiration", "languagetool")
		if st, err := os.Stat(filepath.Join(p, "languagetool-language-modules")); err == nil && st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("no LT data")
		}
		dir = parent
	}
}

func TestCheckUnicodeCasing(t *testing.T) {
	c, err := New(dataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Check("t.txt", "The unicode standard is big.", Options{Language: "en-US"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range res.Findings {
		if f.Rule == "UNICODE_CASING" {
			found = true
			t.Logf("%+v", f)
		}
	}
	if !found {
		t.Fatalf("UNICODE_CASING not found in %#v", res.Findings)
	}
}

func TestCheckWhitespaceAndPattern(t *testing.T) {
	c, err := New(dataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Check("t.txt", "This  has  spaces and unicode too.", Options{Language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	var ws, uni int
	for _, f := range res.Findings {
		switch f.Rule {
		case "WHITESPACE_RULE":
			ws++
		case "UNICODE_CASING":
			uni++
		}
	}
	if ws < 1 {
		t.Fatalf("expected whitespace findings, got %#v", res.Findings)
	}
	if uni < 1 {
		t.Fatalf("expected unicode finding, got %#v", res.Findings)
	}
}
