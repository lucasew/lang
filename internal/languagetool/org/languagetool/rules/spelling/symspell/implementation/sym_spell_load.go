package implementation

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// wordTokenPattern ports SymSpell.parseWords intent (Java "['’\\p{L}-[_]]+"):
// letters, underscore, hyphen, apostrophe (straight/curly); do not split at apostrophes.
// RE2 character-class form of the Java Pattern (find-all on toLowerCase text).
var wordTokenPattern = regexp.MustCompile(`['’\p{L}_\-]+`)

// LoadDictionaryFile ports loadDictionary(String corpus, termIndex, countIndex).
// Returns false if file does not exist (Java).
func (s *SymSpell) LoadDictionaryFile(corpus string, termIndex, countIndex int) bool {
	if s == nil || corpus == "" {
		return false
	}
	f, err := os.Open(corpus)
	if err != nil {
		return false
	}
	defer f.Close()
	return s.LoadDictionary(f, termIndex, countIndex)
}

// LoadDictionary ports loadDictionary(BufferedReader/InputStream, termIndex, countIndex):
// each non-empty line split on \s; key = lineParts[termIndex]; count = parseLong(lineParts[countIndex]);
// createDictionaryEntry into staging; commitStaged.
// Returns true when reader is non-nil and load completes (Java always true after successful open).
func (s *SymSpell) LoadDictionary(r io.Reader, termIndex, countIndex int) bool {
	if s == nil || r == nil {
		return false
	}
	staging := NewSuggestionStage(16384)
	sc := bufio.NewScanner(r)
	// large dictionary lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		// Java: line.split("\\s") — Pattern \s without UNICODE_CHARACTER_CLASS
		// (ASCII WS only; keeps empty mid-fields; drops trailing empties).
		// Not strings.Fields (Unicode WS + collapses consecutive).
		lineParts := javaSplitASCIIWhitespaceSingle(line)
		if len(lineParts) < 2 {
			continue
		}
		if termIndex < 0 || termIndex >= len(lineParts) {
			continue
		}
		if countIndex < 0 || countIndex >= len(lineParts) {
			continue
		}
		key := lineParts[termIndex]
		count, err := strconv.ParseInt(lineParts[countIndex], 10, 64)
		if err != nil {
			// Java: print and continue
			continue
		}
		s.CreateDictionaryEntry(key, count, staging)
	}
	// ignore scanner error like Java catches IOException and prints
	if s.deletes == nil {
		s.deletes = map[int][]string{}
	}
	// ensure capacity-ish: Java new HashMap<>(staging.deleteCount()) when deletes null
	s.CommitStaged(staging)
	return true
}

// javaSplitASCIIWhitespaceSingle ports Java String.split("\\s") without UNICODE_CHARACTER_CLASS:
// delimiter is one of [ \t\n\x0B\f\r]; empty mid-fields kept; trailing empties dropped (limit 0).
func javaSplitASCIIWhitespaceSingle(s string) []string {
	parts := make([]string, 0, 4)
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// CreateDictionaryFile ports createDictionary(String corpus): plain text file,
// parseWords each line, createDictionaryEntry(key, 1, staging), commit.
// Returns false if file does not exist.
func (s *SymSpell) CreateDictionaryFile(corpus string) bool {
	if s == nil || corpus == "" {
		return false
	}
	f, err := os.Open(corpus)
	if err != nil {
		return false
	}
	defer f.Close()
	return s.CreateDictionary(f)
}

// CreateDictionary ports createDictionary from a reader of plain text lines.
func (s *SymSpell) CreateDictionary(r io.Reader) bool {
	if s == nil || r == nil {
		return false
	}
	staging := NewSuggestionStage(16384)
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		for _, key := range parseWords(line) {
			s.CreateDictionaryEntry(key, 1, staging)
		}
	}
	if s.deletes == nil {
		s.deletes = map[int][]string{}
	}
	s.CommitStaged(staging)
	return true
}

// PurgeBelowThresholdWords ports purgeBelowThresholdWords.
func (s *SymSpell) PurgeBelowThresholdWords() {
	if s == nil {
		return
	}
	s.belowThresholdWords = map[string]int64{}
}

// CommitStaged is the Java name for CommitStaging.
func (s *SymSpell) CommitStaged(staging *SuggestionStage) {
	s.CommitStaging(staging)
}

// parseWords ports SymSpell.parseWords:
// Pattern "['’\\p{L}-[_]]+" on text.toLowerCase(); collect find groups.
func parseWords(text string) []string {
	if text == "" {
		return nil
	}
	// Java: text.toLowerCase() then matcher.find
	lower := strings.ToLower(text)
	return wordTokenPattern.FindAllString(lower, -1)
}
