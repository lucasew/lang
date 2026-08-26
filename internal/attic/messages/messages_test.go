package messages

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
			t.Skip("LanguageTool submodule not found")
		}
		dir = parent
	}
}

func TestLoadWhitespaceMessage(t *testing.T) {
	root := repoDataRoot(t)
	b, err := Load(root, "en")
	if err != nil {
		t.Fatal(err)
	}
	got := b.Get("whitespace_repetition")
	want := "Possible typo: you repeated a whitespace"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
