package dumpcheck

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// PlainTextSentenceSource ports org.languagetool.dev.dumpcheck.PlainTextSentenceSource.
type PlainTextSentenceSource struct {
	scanner           *bufio.Scanner
	pending           []string
	articleCount      int
	currentURL        string
	acceptPattern     *regexp.Regexp
	ApplyLengthFilter bool
}

func NewPlainTextSentenceSource(r io.Reader) *PlainTextSentenceSource {
	return &PlainTextSentenceSource{
		scanner:           bufio.NewScanner(r),
		ApplyLengthFilter: true,
	}
}

func (s *PlainTextSentenceSource) GetSource() string { return s.currentURL }

func (s *PlainTextSentenceSource) HasNext() bool {
	s.fill()
	return len(s.pending) > 0
}

func (s *PlainTextSentenceSource) Next() (Sentence, error) {
	s.fill()
	if len(s.pending) == 0 {
		return Sentence{}, fmt.Errorf("no such element")
	}
	line := s.pending[0]
	s.pending = s.pending[1:]
	s.articleCount++
	return NewSentence(line, s.GetSource(), "<plaintext>", "", s.articleCount), nil
}

func (s *PlainTextSentenceSource) fill() {
	for len(s.pending) == 0 && s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# source:") {
			s.currentURL = strings.TrimSpace(strings.TrimPrefix(line, "# source:"))
			// Java uses "# source: " (with space); accept both
			if strings.HasPrefix(line, "# source: ") {
				s.currentURL = line[len("# source: "):]
			}
			continue
		}
		if s.ApplyLengthFilter {
			if !AcceptSentence(line, s.acceptPattern) {
				continue
			}
		}
		s.pending = append(s.pending, line)
	}
}
