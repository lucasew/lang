package en

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Port of org.languagetool.tokenizers.en.TokenizeMultiwordsTest.
// Java only logs warnings for multi-token spelling entries; it does not fail.

var multiwordsFilesToTest = []string{
	"/en/added.txt", "/en/removed.txt",
	"/en/hunspell/ignore.txt", "/en/hunspell/prohibit.txt", "/en/hunspell/prohibit_custom.txt",
	"/en/hunspell/spelling.txt", "/en/hunspell/spelling_custom.txt",
	"/en/hunspell/spelling_en-AU.txt", "/en/hunspell/spelling_en-CA.txt",
	"/en/hunspell/spelling_en-GB.txt", "/en/hunspell/spelling_en-NZ.txt",
	"/en/hunspell/spelling_en-US.txt", "/en/hunspell/spelling_en-ZA.txt",
	"/en/hunspell/spelling_merged.txt",
}

func TestTokenizeMultiwords_Tokenize(t *testing.T) {
	root := findModuleRoot(t)
	resBase := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource")
	// multiwords may live under en resources
	mwPath := filepath.Join(resBase, "en/multiwords.txt")
	if _, err := os.Stat(mwPath); err != nil {
		// try core resources layout
		mwPath = filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/multiwords.txt")
	}
	if _, err := os.Stat(mwPath); err != nil {
		t.Logf("multiwords.txt not found, skipping resource scan: %v", err)
		return
	}
	multiwords, err := loadWordColumn(mwPath)
	if err != nil {
		t.Fatal(err)
	}
	mwSet := map[string]bool{}
	for _, w := range multiwords {
		mwSet[strings.ReplaceAll(w, "’", "'")] = true
	}
	wt := NewEnglishWordTokenizer()
	for _, fileName := range multiwordsFilesToTest {
		p := filepath.Join(resBase, strings.TrimPrefix(fileName, "/"))
		if _, err := os.Stat(p); err != nil {
			t.Logf("missing %s: %v", fileName, err)
			continue
		}
		wordList, err := loadWordColumn(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, word := range wordList {
			if mwSet[strings.ReplaceAll(word, "’", "'")] {
				continue
			}
			tokens := wt.Tokenize(word)
			tokensBySpace := strings.Split(word, " ")
			nonSpace := filterSpaces(tokens)
			if len(tokens) > 1 && !stringSlicesEqual(nonSpace, tokensBySpace) {
				t.Logf("WARNING: %q in %q is multi-token - please make sure it actually works.", word, fileName)
			}
		}
	}
}

func filterSpaces(toks []string) []string {
	var out []string
	for _, k := range toks {
		if k != " " {
			out = append(out, k)
		}
	}
	return out
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func loadWordColumn(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "\t")
		lines = append(lines, parts[0])
	}
	return lines, sc.Err()
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
