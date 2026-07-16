// Package tagger implements LanguageTool-style POS tagging via morfologik.
package tagger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/morfologik"
)

// Reading is one stem+tag for a token.
type Reading struct {
	Lemma string
	POS   string
}

// Tagger tags words for a language family.
type Tagger struct {
	dict  *morfologik.Dictionary
	added map[string][]Reading
	mu    sync.RWMutex
}

// OpenEnglish loads english.dict + added.txt.
func OpenEnglish(dataRoot string) (*Tagger, error) {
	dictPath, err := findResource(dataRoot, "en", "english.dict")
	if err != nil {
		return nil, err
	}
	d, err := morfologik.OpenDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	t := &Tagger{dict: d, added: map[string][]Reading{}}
	// added.txt lives in LT module tree (text), not necessarily next to binary dict
	for _, base := range resourceBases(dataRoot, "en") {
		for _, name := range []string{"added.txt", "added_custom.txt"} {
			_ = t.loadAdded(filepath.Join(base, name))
		}
	}
	return t, nil
}

func resourceBases(dataRoot, family string) []string {
	var out []string
	out = append(out, filepath.Join(dataRoot, "languagetool-language-modules", family, "src", "main", "resources", "org", "languagetool", "resource", family))
	// third_party/english-pos-dict
	if p := findUp("third_party/english-pos-dict/org/languagetool/resource/" + family); p != "" {
		out = append(out, p)
	}
	return out
}

func findResource(dataRoot, family, name string) (string, error) {
	candidates := []string{
		filepath.Join(dataRoot, "languagetool-language-modules", family, "src", "main", "resources", "org", "languagetool", "resource", family, name),
	}
	if p := findUp(filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", family, name)); p != "" {
		candidates = append([]string{p}, candidates...)
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("resource %s/%s not found (run scripts/fetch-english-dicts.sh)", family, name)
}

func findUp(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil {
			_ = st
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func (t *Tagger) loadAdded(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		form, lemma, pos := parts[0], parts[1], parts[2]
		t.added[form] = append(t.added[form], Reading{Lemma: lemma, POS: pos})
	}
	return sc.Err()
}

// TagWord returns readings for a surface token (English rules).
func (t *Tagger) TagWord(word string) []Reading {
	if word == "" {
		return nil
	}
	if word == "SENT_START" {
		return []Reading{{Lemma: "SENT_START", POS: "SENT_START"}}
	}
	w := strings.ReplaceAll(word, "’", "'")
	var out []Reading
	seen := map[string]bool{}
	add := func(rs []Reading) {
		for _, r := range rs {
			key := r.Lemma + "\t" + r.POS
			if !seen[key] {
				seen[key] = true
				out = append(out, r)
			}
		}
	}

	add(t.lookupSurface(w))
	lower := strings.ToLower(w)
	isLower := w == lower
	isMixed := isMixedCase(w)
	isAllUpper := isAllUppercase(w)
	if !isLower && !isMixed {
		add(t.lookupSurface(lower))
	}
	if len(out) == 0 && isAllUpper {
		add(t.lookupSurface(firstUpper(lower)))
	}
	t.mu.RLock()
	if rs, ok := t.added[w]; ok {
		add(rs)
	}
	if rs, ok := t.added[lower]; ok && !isLower {
		add(rs)
	}
	t.mu.RUnlock()
	return out
}

func (t *Tagger) lookupSurface(w string) []Reading {
	forms, err := t.dict.Lookup(w)
	if err != nil || len(forms) == 0 {
		return nil
	}
	out := make([]Reading, 0, len(forms))
	for _, f := range forms {
		out = append(out, Reading{Lemma: f.Stem, POS: f.Tag})
	}
	return out
}

func isMixedCase(s string) bool {
	hasUpper, hasLower := false, false
	first := true
	firstWasUpper := false
	for _, r := range s {
		if first {
			firstWasUpper = unicode.IsUpper(r)
			first = false
		}
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}
	if !hasUpper || !hasLower {
		return false
	}
	if firstWasUpper {
		allRestLower := true
		i := 0
		for _, r := range s {
			if i == 0 {
				i++
				continue
			}
			if unicode.IsUpper(r) {
				allRestLower = false
				break
			}
			i++
		}
		if allRestLower {
			return false
		}
	}
	return true
}

func isAllUppercase(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func firstUpper(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
