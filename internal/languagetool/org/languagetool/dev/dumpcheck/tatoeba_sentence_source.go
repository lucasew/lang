package dumpcheck

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// TatoebaSentenceSource ports org.languagetool.dev.dumpcheck.TatoebaSentenceSource.
// Input is tab-separated: id \t lang \t sentence (one language already filtered).
type TatoebaSentenceSource struct {
	scanner       *bufio.Scanner
	pending       []Sentence
	articleCount  int
	acceptPattern *regexp.Regexp
	// ApplyLengthFilter when false accepts any well-formed 3-column line (Java
	// still applies acceptSentence; tests with short fixtures may need true).
	ApplyLengthFilter bool
}

func NewTatoebaSentenceSource(r io.Reader) *TatoebaSentenceSource {
	return &TatoebaSentenceSource{
		scanner:           bufio.NewScanner(r),
		ApplyLengthFilter: true,
	}
}

func NewTatoebaSentenceSourceWithFilter(r io.Reader, filter *regexp.Regexp) *TatoebaSentenceSource {
	s := NewTatoebaSentenceSource(r)
	s.acceptPattern = filter
	return s
}

func (s *TatoebaSentenceSource) GetSource() string { return "tatoeba" }

func (s *TatoebaSentenceSource) HasNext() bool {
	s.fill()
	return len(s.pending) > 0
}

func (s *TatoebaSentenceSource) Next() (Sentence, error) {
	s.fill()
	if len(s.pending) == 0 {
		return Sentence{}, fmt.Errorf("no such element")
	}
	out := s.pending[0]
	s.pending = s.pending[1:]
	return out, nil
}

func (s *TatoebaSentenceSource) fill() {
	for len(s.pending) == 0 && s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			// Java: skip unexpected format
			continue
		}
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		sentence := parts[2]
		if s.ApplyLengthFilter {
			if !AcceptSentence(sentence, s.acceptPattern) {
				continue
			}
		} else if s.acceptPattern != nil && !s.acceptPattern.MatchString(sentence) {
			continue
		}
		s.articleCount++
		s.pending = append(s.pending, NewSentence(
			sentence,
			s.GetSource(),
			fmt.Sprintf("Tatoeba-%d", id),
			"http://tatoeba.org",
			s.articleCount,
		))
	}
}
