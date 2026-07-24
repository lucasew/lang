package uk

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/multiwords.txt
var multiwordsFS embed.FS

var (
	multiwordsOnce  sync.Once
	multiwordsLines []string

	ukMultiwordChunkerOnce sync.Once
	ukMultiwordChunkerInst *disambiguation.MultiWordChunker2
)

// UkrainianMultiwordChunker ports tagging.disambiguation.uk.UkrainianMultiwordChunker
// (MultiWordChunker2 + /POS-regex matchText).
//
// Java:
//
//	class UkrainianMultiwordChunker extends MultiWordChunker2 {
//	  UkrainianMultiwordChunker(String filename, boolean allowFirstCapitalized) {
//	    super(filename, allowFirstCapitalized);
//	  }
//	  protected boolean matches(String matchText, AnalyzedTokenReadings inputTokens) {
//	    if (!matchText.startsWith("/")) return super.matches(...);
//	    return PosTagHelper.hasPosTag(inputTokens, Pattern.compile(matchText.substring(1)));
//	  }
//	}
//
// Hybrid: new UkrainianMultiwordChunker("/uk/multiwords.txt", true)
type UkrainianMultiwordChunker = disambiguation.MultiWordChunker2

// NewUkrainianMultiwordChunker builds from phrase\ttag lines (allowFirstCapitalized=true).
func NewUkrainianMultiwordChunker(lines []string) *disambiguation.MultiWordChunker2 {
	c := disambiguation.NewMultiWordChunker2(lines, true)
	c.MatchesFn = ukMultiwordMatches
	return c
}

// NewDefaultUkrainianMultiwordChunker loads official /uk/multiwords.txt
// (discovered inspiration path when present; embedded identical copy as fallback).
func NewDefaultUkrainianMultiwordChunker() *disambiguation.MultiWordChunker2 {
	return NewUkrainianMultiwordChunker(LoadUkrainianMultiwordsLines())
}

// UkrainianMultiwordChunkerDefault returns a process-cached MultiWordChunker2 for
// official /uk/multiwords.txt (Java UkrainianHybridDisambiguator.chunker field).
// Prefer discoverable official file; embedded fallback when discovery fails.
func UkrainianMultiwordChunkerDefault() *disambiguation.MultiWordChunker2 {
	ukMultiwordChunkerOnce.Do(func() {
		ukMultiwordChunkerInst = NewDefaultUkrainianMultiwordChunker()
	})
	return ukMultiwordChunkerInst
}

// LoadUkrainianMultiwordsLines returns phrase\ttag lines from official multiwords.txt.
// Prefer DiscoverUkrainianMultiwords(); fall back to embedded data/multiwords.txt
// (byte-identical to official resource when submodule is present).
func LoadUkrainianMultiwordsLines() []string {
	multiwordsOnce.Do(func() {
		if p := DiscoverUkrainianMultiwords(); p != "" {
			if f, err := os.Open(p); err == nil {
				lines, err := parseUkrainianMultiwordReader(f)
				_ = f.Close()
				if err == nil {
					multiwordsLines = lines
					return
				}
			}
		}
		f, err := multiwordsFS.Open("data/multiwords.txt")
		if err != nil {
			return
		}
		defer f.Close()
		lines, err := parseUkrainianMultiwordReader(f)
		if err != nil {
			return
		}
		multiwordsLines = lines
	})
	return multiwordsLines
}

// NewUkrainianMultiwordChunkerFromReader loads multiwords lines then builds the chunker.
// Input must be Java MultiWordChunker2 format: tab-separated phrase\ttag (no invent glue).
func NewUkrainianMultiwordChunkerFromReader(r io.Reader) (*disambiguation.MultiWordChunker2, error) {
	lines, err := parseUkrainianMultiwordReader(r)
	if err != nil {
		return nil, err
	}
	return NewUkrainianMultiwordChunker(lines), nil
}

// NewUkrainianMultiwordChunkerFromPath opens multiwords file path.
func NewUkrainianMultiwordChunkerFromPath(path string) (*disambiguation.MultiWordChunker2, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewUkrainianMultiwordChunkerFromReader(f)
}

// DiscoverUkrainianMultiwords finds official uk/multiwords.txt
// (Java resource /uk/multiwords.txt used by UkrainianHybridDisambiguator.chunker).
func DiscoverUkrainianMultiwords() string {
	if p := os.Getenv("LANG_UK_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
		"src", "main", "resources", "org", "languagetool", "resource", "uk", "multiwords.txt")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// parseUkrainianMultiwordReader ports MultiWordChunker2.loadWords:
// trim, skip empty and # lines; remaining lines must be tab-separated phrase\ttag
// (Java throws on non-tab format — no glued-POS invent path).
func parseUkrainianMultiwordReader(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Java MultiWordChunker2: split("\t") and require length == 2
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Invalid format in multiwords: '%s', expected two tab-separated parts", line)
		}
		if tools.JavaStringTrim(parts[0]) == "" || tools.JavaStringTrim(parts[1]) == "" {
			return nil, fmt.Errorf("Invalid format in multiwords: '%s', empty phrase or tag", line)
		}
		// Keep line as Java loadWords does (trimmed whole line, not re-joined parts).
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

// ukMultiwordMatches ports UkrainianMultiwordChunker.matches:
// non-/ → token equality; /pattern → PosTagHelper.hasPosTag full-match regex on readings.
func ukMultiwordMatches(matchText string, tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	if !strings.HasPrefix(matchText, "/") {
		return matchText == tok.GetToken()
	}
	// Java: Pattern.compile(matchText.substring(1)); matcher(posTag).matches()
	re, err := regexp.Compile(matchText[1:])
	if err != nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		loc := re.FindStringIndex(pos)
		if loc != nil && loc[0] == 0 && loc[1] == len(pos) {
			return true
		}
	}
	return false
}
