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
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// plainTextAcceptCache caches loaded accept words by absolute file path.
var plainTextAcceptCache sync.Map // string -> []string

// plainRuntimeDictCache ports Java MorfologikMultiSpeller.dicPathToDict:
// FSABuilder + CFSA2Serializer for plain-text spelling lists is expensive (~1s for PT),
// so reuse Dictionary across speller1/2/3 and rule constructions (not user dicts).
var plainRuntimeDictCache sync.Map // string -> *atticmorfo.Dictionary

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
//
// cacheKey non-empty enables dicPathToDict-style reuse (not for user dicts).
// Java: cacheKey = plainTextPaths + "|" + infoPath.
func (s *MorfologikSpeller) AttachWordsAsBinaryFSA(words []string, infoBesideDictPath string) int {
	return s.attachWordsAsBinaryFSA(words, infoBesideDictPath, "")
}

// AttachWordsAsBinaryFSACached is AttachWordsAsBinaryFSA with Java dicPathToDict key.
func (s *MorfologikSpeller) AttachWordsAsBinaryFSACached(words []string, infoBesideDictPath, cacheKey string) int {
	return s.attachWordsAsBinaryFSA(words, infoBesideDictPath, cacheKey)
}

func (s *MorfologikSpeller) attachWordsAsBinaryFSA(words []string, infoBesideDictPath, cacheKey string) int {
	if s == nil || len(words) == 0 {
		return 0
	}
	// Load .info flags (Java uses dictPath.replace(.dict, .info) for plain-text runtime dict).
	var info map[string]string
	infoPath := ""
	if infoBesideDictPath != "" {
		infoPath = strings.TrimSuffix(infoBesideDictPath, filepath.Ext(infoBesideDictPath)) + ".info"
		if m, err := readSpellerInfoFile(infoPath); err == nil {
			info = m
			s.ApplyInfoProperties(m)
		}
	}
	// Dedupe empty only — Java FSABuilder gets the line list (may include duplicates).
	uniq := make([]string, 0, len(words))
	for _, w := range words {
		w = tools.JavaStringTrim(w)
		if w == "" {
			continue
		}
		uniq = append(uniq, w)
	}
	if len(uniq) == 0 {
		return 0
	}

	var d *atticmorfo.Dictionary
	if cacheKey != "" {
		if v, ok := plainRuntimeDictCache.Load(cacheKey); ok {
			if cached, ok := v.(*atticmorfo.Dictionary); ok && cached != nil {
				d = cached
			}
		}
	}
	if d == nil {
		// Creating the dictionary at runtime can take ~1s for large spelling.txt merges
		// (Java MorfologikMultiSpeller comment ~50ms for 50KB; PT multiwords is larger).
		d = atticmorfo.NewDictionaryFromWords(uniq, info)
		if d == nil {
			return 0
		}
		if cacheKey != "" {
			actual, _ := plainRuntimeDictCache.LoadOrStore(cacheKey, d)
			if cached, ok := actual.(*atticmorfo.Dictionary); ok && cached != nil {
				d = cached
			}
		}
	}

	// Wire like AttachBinaryDictionary: per-instance Speller (sticky containsSeparators).
	// Plain FSABuilder words are ExactMatch → first isInDictionary clears separators (Java).
	// Do not populate Words map: Java MorfologikSpeller holds Dictionary only (membership via FSA).
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
		rel = strings.TrimPrefix(tools.JavaStringTrim(rel), "/")
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
			// Java: line.split("#")[0].trim()
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = line[:i]
			}
			line = tools.JavaStringTrim(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			out = append(out, line)
		}
	}
	return out
}
