package uk

import (
	"bufio"
	"embed"
	"io"
	"os"
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

// NewDefaultUkrainianMultiwordChunker loads official /uk/multiwords.txt (embedded).
func NewDefaultUkrainianMultiwordChunker() *disambiguation.MultiWordChunker2 {
	return NewUkrainianMultiwordChunker(LoadUkrainianMultiwordsLines())
}

// LoadUkrainianMultiwordsLines returns phrase\ttag lines from embedded multiwords.txt.
func LoadUkrainianMultiwordsLines() []string {
	multiwordsOnce.Do(func() {
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

// parseUkrainianMultiwordReader accepts:
//   - standard phrase\ttag lines (Java MultiWordChunker2)
//   - official UK resource lines with glued POS suffix (а капелаadv, без сумнівуinsert)
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
		if i := strings.IndexByte(line, '#'); i >= 0 {
			// only strip trailing comments when space before #
			if i == 0 || line[i-1] == ' ' {
				line = tools.JavaStringTrim(line[:i])
			}
		}
		if line == "" {
			continue
		}
		if norm, ok := normalizeUkrainianMultiwordLine(line); ok {
			lines = append(lines, norm)
		}
	}
	return lines, sc.Err()
}

// glued POS suffixes used in official UK multiwords.txt (longest first).
var ukMultiwordGluedTags = []string{
	"insert",
	"adv",
}

func normalizeUkrainianMultiwordLine(line string) (string, bool) {
	if strings.Contains(line, "\t") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || tools.JavaStringTrim(parts[0]) == "" || tools.JavaStringTrim(parts[1]) == "" {
			return "", false
		}
		return tools.JavaStringTrim(parts[0]) + "\t" + tools.JavaStringTrim(parts[1]), true
	}
	// glued tag at end (no tab in official resource)
	for _, tag := range ukMultiwordGluedTags {
		if strings.HasSuffix(line, tag) {
			phrase := tools.JavaStringTrim(line[:len(line)-len(tag)])
			if phrase == "" {
				continue
			}
			// phrase must contain a space or hyphen multi-token, or single token is ok
			return phrase + "\t" + tag, true
		}
	}
	// /POS-regex style match tokens may appear as last field after space — unsupported without tab
	return "", false
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
		loc := re.FindStringIndex(pos)
		if loc != nil && loc[0] == 0 && loc[1] == len(pos) {
			return true
		}
	}
	return false
}
