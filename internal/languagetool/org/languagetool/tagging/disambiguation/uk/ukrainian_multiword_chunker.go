package uk

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// UkrainianMultiwordChunker ports tagging.disambiguation.uk.UkrainianMultiwordChunker
// (MultiWordChunker2 + /POS-regex matchText).
type UkrainianMultiwordChunker = disambiguation.MultiWordChunker2

// NewUkrainianMultiwordChunker builds from phrase\ttag lines (allowFirstCapitalized=true).
func NewUkrainianMultiwordChunker(lines []string) *disambiguation.MultiWordChunker2 {
	c := disambiguation.NewMultiWordChunker2(lines, true)
	c.MatchesFn = ukMultiwordMatches
	return c
}

// NewUkrainianMultiwordChunkerFromReader loads multiwords.txt lines then builds the chunker.
func NewUkrainianMultiwordChunkerFromReader(r io.Reader) (*disambiguation.MultiWordChunker2, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
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

// ukMultiwordMatches ports UkrainianMultiwordChunker.matches:
// non-/ → token equality; /pattern → POS tag full-match regex on readings.
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
		// Matcher.matches() = entire string
		loc := re.FindStringIndex(pos)
		if loc != nil && loc[0] == 0 && loc[1] == len(pos) {
			return true
		}
	}
	return false
}
