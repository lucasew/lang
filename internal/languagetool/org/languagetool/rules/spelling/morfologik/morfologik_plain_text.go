package morfologik

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// plainTextAcceptCache caches loaded accept words by absolute file path.
var plainTextAcceptCache sync.Map // string -> []string

// PrepareLineFn ports Language.prepareLineForSpeller for plain-text multi-speller lines.
// Nil → treat line as raw surface (strip # comments only).
type PrepareLineFn func(line string) []string

// LoadPlainTextAcceptFile loads a Java multi-speller plain-text .txt (spelling.txt, multiwords.txt)
// into the map Words set. prepareLine nil uses default strip (# comment, trim).
// Missing file is skipped (fail closed).
func (s *MorfologikSpeller) LoadPlainTextAcceptFile(path string, prepareLine PrepareLineFn) int {
	if s == nil || path == "" {
		return 0
	}
	words := loadPlainTextAcceptCached(path, prepareLine)
	n := 0
	for _, w := range words {
		if w == "" {
			continue
		}
		// Multi-token phrases are multi-speller entries; single-token Match uses token surface.
		// Still register all for GetSuggestions / membership parity.
		s.AddWord(w)
		n++
	}
	return n
}

// LoadPlainTextAcceptClasspaths discovers resource-dir relative paths and loads accept words.
// Returns total words added (may re-add duplicates).
func (s *MorfologikSpeller) LoadPlainTextAcceptClasspaths(relPaths []string, prepareLine PrepareLineFn) int {
	if s == nil {
		return 0
	}
	total := 0
	for _, rel := range relPaths {
		rel = strings.TrimPrefix(strings.TrimSpace(rel), "/")
		if rel == "" {
			continue
		}
		p := spelling.DiscoverSpellingResource(rel)
		if p == "" {
			continue
		}
		total += s.LoadPlainTextAcceptFile(p, prepareLine)
	}
	return total
}

func loadPlainTextAcceptCached(path string, prepareLine PrepareLineFn) []string {
	key := path
	if prepareLine != nil {
		key = path + "|prep"
	}
	if v, ok := plainTextAcceptCache.Load(key); ok {
		if ws, ok := v.([]string); ok {
			return ws
		}
	}
	ws := loadPlainTextAcceptFile(path, prepareLine)
	plainTextAcceptCache.Store(key, ws)
	return ws
}

func loadPlainTextAcceptFile(path string, prepareLine PrepareLineFn) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	// multiwords can be long lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		original := sc.Text()
		var lines []string
		if prepareLine != nil {
			lines = prepareLine(original)
		} else {
			lines = []string{original}
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			if line == "" {
				continue
			}
			out = append(out, line)
		}
	}
	return out
}
