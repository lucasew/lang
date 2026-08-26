package dumpcheck

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// CommonCrawlSentenceSource ports org.languagetool.dev.dumpcheck.CommonCrawlSentenceSource
// for already-decompressed line-oriented input (xz auto-decode deferred).
type CommonCrawlSentenceSource struct {
	scanner       *bufio.Scanner
	pending       []string
	articleCount  int
	acceptPattern *regexp.Regexp
	// Counters for diagnostics (Java fields).
	TooShort, TooLong, Empty, WrongStartChar, WrongEndChar int
}

const (
	ccMinLength = 15
	ccMaxLength = 250
)

func NewCommonCrawlSentenceSource(r io.Reader) *CommonCrawlSentenceSource {
	return &CommonCrawlSentenceSource{scanner: bufio.NewScanner(r)}
}

func NewCommonCrawlSentenceSourceWithFilter(r io.Reader, filter *regexp.Regexp) *CommonCrawlSentenceSource {
	s := NewCommonCrawlSentenceSource(r)
	s.acceptPattern = filter
	return s
}

func (s *CommonCrawlSentenceSource) GetSource() string { return "commoncrawl" }

func (s *CommonCrawlSentenceSource) HasNext() bool {
	s.fill()
	return len(s.pending) > 0
}

func (s *CommonCrawlSentenceSource) Next() (Sentence, error) {
	s.fill()
	if len(s.pending) == 0 {
		return Sentence{}, fmt.Errorf("no such element")
	}
	line := s.pending[0]
	s.pending = s.pending[1:]
	s.articleCount++
	return NewSentence(line, s.GetSource(), "", "", s.articleCount), nil
}

func (s *CommonCrawlSentenceSource) fill() {
	for len(s.pending) == 0 && s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			s.Empty++
			continue
		}
		if len(line) < ccMinLength {
			s.TooShort++
			continue
		}
		if len(line) > ccMaxLength {
			s.TooLong++
			continue
		}
		r0 := []rune(line)[0]
		if !unicode.IsUpper(r0) {
			s.WrongStartChar++
			continue
		}
		last := []rune(line)[len([]rune(line))-1]
		if last != '.' && last != '!' && last != '?' && last != '…' {
			s.WrongEndChar++
			continue
		}
		if s.acceptPattern != nil && !s.acceptPattern.MatchString(line) {
			continue
		}
		// also apply generic length/token filter used by other sources
		if !AcceptSentence(line, nil) {
			continue
		}
		s.pending = append(s.pending, line)
	}
}
