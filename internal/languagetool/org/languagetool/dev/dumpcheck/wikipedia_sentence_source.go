package dumpcheck

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/wikipedia"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// WikipediaSentenceSource ports org.languagetool.dev.dumpcheck.WikipediaSentenceSource
// using encoding/xml and SimpleWikipediaTextFilter (Sweble deferred).
type WikipediaSentenceSource struct {
	decoder       *xml.Decoder
	pending       []Sentence
	articleCount  int
	langCode      string
	acceptPattern *regexp.Regexp
	filter        *wikipedia.SimpleWikipediaTextFilter
	sentTok       *tokenizers.EnglishSRXSentenceTokenizer
}

func NewWikipediaSentenceSource(r io.Reader, langCode string) *WikipediaSentenceSource {
	if langCode == "" {
		langCode = "en"
	}
	return &WikipediaSentenceSource{
		decoder:  xml.NewDecoder(r),
		langCode: langCode,
		filter:   wikipedia.NewSimpleWikipediaTextFilter(),
		sentTok:  tokenizers.NewEnglishSRXSentenceTokenizer(),
	}
}

func (s *WikipediaSentenceSource) GetSource() string { return "wikipedia" }

func (s *WikipediaSentenceSource) HasNext() bool {
	s.fill()
	return len(s.pending) > 0
}

func (s *WikipediaSentenceSource) Next() (Sentence, error) {
	s.fill()
	if len(s.pending) == 0 {
		return Sentence{}, fmt.Errorf("no such element")
	}
	out := s.pending[0]
	s.pending = s.pending[1:]
	return out, nil
}

func (s *WikipediaSentenceSource) fill() {
	for len(s.pending) == 0 {
		tok, err := s.decoder.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "page" {
			continue
		}
		var page wikiPage
		if err := s.decoder.DecodeElement(&page, &se); err != nil {
			continue
		}
		s.articleCount++
		text := page.Revision.Text
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(text)), "#redirect") {
			continue
		}
		plain := s.filter.Filter(text)
		for _, sent := range s.sentTok.Tokenize(plain) {
			if !AcceptSentence(sent, s.acceptPattern) {
				continue
			}
			titleWithID := fmt.Sprintf("%s/%d", page.Title, hash32(sent))
			url := fmt.Sprintf("http://%s.wikipedia.org/wiki/%s", s.langCode, page.Title)
			s.pending = append(s.pending, NewSentence(sent, s.GetSource(), titleWithID, url, s.articleCount))
		}
	}
}

type wikiPage struct {
	Title    string       `xml:"title"`
	NS       string       `xml:"ns"`
	Revision wikiRevision `xml:"revision"`
}

type wikiRevision struct {
	Text string `xml:"text"`
}

func hash32(s string) int32 {
	// stable non-cryptographic hash for title suffix (Java uses String.hashCode)
	var h int32
	for _, r := range s {
		h = 31*h + int32(r)
	}
	return h
}
