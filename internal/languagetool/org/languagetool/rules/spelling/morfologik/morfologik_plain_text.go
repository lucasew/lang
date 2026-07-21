package morfologik

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
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
// Prefer AttachWordsAsBinaryFSA after collecting all lines (Java FSABuilder runtime dict).
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

// AttachWordsAsBinaryFSA ports Java MorfologikMultiSpeller.getDictionary:
// FSABuilder.build(sorted UTF-8 lines) + Dictionary.read with binary .info metadata.
// Replaces invent map-only SpellerED scan with real SpellerFSA findRepl over the word FSA.
// Returns number of words encoded (0 if empty).
func (s *MorfologikSpeller) AttachWordsAsBinaryFSA(words []string, infoBesideDictPath string) int {
	if s == nil || len(words) == 0 {
		return 0
	}
	// Load .info flags (Java uses dictPath.replace(.dict, .info) for plain-text runtime dict).
	var info map[string]string
	if infoBesideDictPath != "" {
		infoPath := strings.TrimSuffix(infoBesideDictPath, filepath.Ext(infoBesideDictPath)) + ".info"
		if m, err := readSpellerInfoFile(infoPath); err == nil {
			info = m
			s.ApplyInfoProperties(m)
		}
	}
	// Also keep Words for HasDictionary / any map fallback.
	uniq := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		s.AddWord(w)
		uniq = append(uniq, w)
	}
	if len(uniq) == 0 {
		return 0
	}
	d := atticmorfo.NewDictionaryFromWords(uniq, info)
	if d == nil {
		return 0
	}
	// Wire like AttachBinaryDictionary: per-instance Speller (sticky containsSeparators).
	// Plain FSABuilder words are ExactMatch → first isInDictionary clears separators (Java).
	s.syncDictSpellerMeta(d)
	sp := atticmorfo.NewSpeller(d, s.MaxEditDistance)
	sp.SyncFromDict()
	s.binarySpeller = sp
	s.InDictionaryFn = sp.IsInDictionary
	s.binaryDict = d
	s.BinaryDictPath = s.FileInClassPath
	s.FrequencyIncluded = d.FrequencyIncluded()
	s.GetFrequencyFn = func(word string) int {
		return d.GetFrequency(word)
	}
	// Full candidate list (Java getSuggestions has no 8-cap); GetWeightedSuggestions prefers binaryDict path.
	s.WeightedSuggestFn = func(word string) []WeightedSuggestion {
		return s.binaryFindReplacementCandidates(d, word, 0)
	}
	s.SuggestFn = func(word string) []string {
		return wordsFromWeighted(s.binaryFindReplacementCandidates(d, word, 0))
	}
	return len(uniq)
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
	// Java MorfologikMultiSpeller.getLines uses Language.prepareLineForSpeller —
	// EN vs ES filters on the same spelling_global.txt yield different word sets.
	// Cache key must include prepare identity (function pointer), not a shared "|prep".
	key := path
	if prepareLine != nil {
		key = fmt.Sprintf("%s|prep:%p", path, prepareLine)
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
