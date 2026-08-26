package errorcorpus

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var simpleCorpusLineRE = regexp.MustCompile(`^\d+\.`)

// SimpleCorpus ports org.languagetool.dev.errorcorpus.SimpleCorpus.
// Format: "1. This is _a_ error. => an"
type SimpleCorpus struct {
	lines []string
	pos   int
}

// NewSimpleCorpus loads numbered lines from a single file.
func NewSimpleCorpus(path string) (*SimpleCorpus, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if simpleCorpusLineRE.MatchString(line) {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return &SimpleCorpus{lines: lines}, nil
}

func (c *SimpleCorpus) HasNext() bool { return c != nil && c.pos < len(c.lines) }

func (c *SimpleCorpus) Next() (*ErrorSentence, error) {
	if !c.HasNext() {
		return nil, fmt.Errorf("no such element")
	}
	line := c.lines[c.pos]
	c.pos++
	return parseSimpleCorpusLine(line)
}

func (c *SimpleCorpus) Len() int {
	if c == nil {
		return 0
	}
	return len(c.lines)
}

func parseSimpleCorpusLine(line string) (*ErrorSentence, error) {
	normalized := regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(line, "")
	normalizedNoCorrection := strings.TrimSpace(regexp.MustCompile(`=>.*`).ReplaceAllString(normalized, ""))
	startError := strings.Index(normalized, "_")
	if startError < 0 {
		return nil, fmt.Errorf("no '_..._' marker found: %s", line)
	}
	endError := strings.Index(normalized[startError+1:], "_")
	if endError < 0 {
		return nil, fmt.Errorf("no '_..._' marker found: %s", line)
	}
	endError = startError + 1 + endError
	startCorrectionMarker := strings.Index(normalized, "=>")
	if startCorrectionMarker < 0 {
		return nil, fmt.Errorf("no '=>' marker found: %s", line)
	}
	correction := strings.TrimSpace(normalized[startCorrectionMarker+len("=>"):])
	// Java: new Error(startError + 1, endError - 1, correction) — markup positions with underscores
	errors := []Error{{
		StartPos:   startError + 1,
		EndPos:     endError - 1,
		Correction: correction,
	}}
	// plain: underscores → spaces, collapse
	plain := strings.ReplaceAll(normalizedNoCorrection, "_", " ")
	plain = regexp.MustCompile(`\s+`).ReplaceAllString(plain, " ")
	plain = strings.TrimSpace(plain)
	return &ErrorSentence{
		MarkupText: normalizedNoCorrection,
		PlainText:  plain,
		Errors:     errors,
	}, nil
}
