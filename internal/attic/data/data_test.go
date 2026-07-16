package data

import (
	"os"
	"path/filepath"
	"testing"
)

func repoDataRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		candidate := filepath.Join(dir, "inspiration", "languagetool")
		if st, err := os.Stat(filepath.Join(candidate, "languagetool-language-modules")); err == nil && st.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("LanguageTool submodule not found from ", wd)
		}
		dir = parent
	}
}

func TestResolveFlag(t *testing.T) {
	root := repoDataRoot(t)
	got, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("empty root")
	}
}

func TestDiscoverLanguages(t *testing.T) {
	root := repoDataRoot(t)
	langs, err := DiscoverLanguages(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(langs) < 10 {
		t.Fatalf("expected many languages, got %d", len(langs))
	}
	if _, ok := Lookup(langs, "en-US"); !ok {
		t.Fatal("en-US not found")
	}
	if _, ok := Lookup(langs, "en"); !ok {
		t.Fatal("en not found")
	}
}
