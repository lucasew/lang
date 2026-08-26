package wikipedia

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

var (
	fullIPAPattern = regexp.MustCompile(`'''?(.*?)'''?\s+\[?\{\{IPA\|([^}]*)\}\}`)
	ipaPattern     = regexp.MustCompile(`\{\{IPA\|([^}]*)\}\}`)
)

// IPAHit is one extracted IPA entry.
type IPAHit struct {
	Title string
	Word  string // may be empty when only bare {{IPA|...}} matched
	IPA   string
}

// IpaExtractor ports org.languagetool.dev.wikipedia.IpaExtractor (regex green slice).
type IpaExtractor struct {
	ArticleCount int
	IPACount     int
	Hits         []IPAHit
}

func NewIpaExtractor() *IpaExtractor { return &IpaExtractor{} }

// ExtractFromText applies the Java FULL_IPA / IPA patterns to wikitext.
// Returns number of IPA hits found (0 or 1 per article, same as Java).
func (e *IpaExtractor) ExtractFromText(title, wikiText string) int {
	if e == nil {
		return 0
	}
	if !strings.Contains(wikiText, "{{IPA") {
		return 0
	}
	if m := fullIPAPattern.FindStringSubmatch(wikiText); m != nil {
		e.Hits = append(e.Hits, IPAHit{Title: title, Word: m[1], IPA: m[2]})
		e.IPACount++
		return 1
	}
	if m := ipaPattern.FindStringSubmatch(wikiText); m != nil {
		e.Hits = append(e.Hits, IPAHit{Title: title, Word: "", IPA: m[1]})
		e.IPACount++
		return 1
	}
	return 0
}

// ExtractFromMediaWikiXML walks <page><title>/<text> like the Java dump loop.
func (e *IpaExtractor) ExtractFromMediaWikiXML(r io.Reader) error {
	dec := xml.NewDecoder(r)
	var title string
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "title":
			var t string
			if err := dec.DecodeElement(&t, &se); err == nil {
				title = t
				e.ArticleCount++
			}
		case "text":
			var text string
			if err := dec.DecodeElement(&text, &se); err == nil {
				e.ExtractFromText(title, text)
			}
		}
	}
}
