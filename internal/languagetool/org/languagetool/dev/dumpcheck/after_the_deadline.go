package dumpcheck

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// AtDMatch is one After-the-Deadline error entry (description: string).
type AtDMatch struct {
	Description string
	String      string
}

func (m AtDMatch) Format() string {
	return m.Description + ": " + m.String
}

// ParseAtDResultXML ports AfterTheDeadlineChecker.getMatches (//error string+description).
func ParseAtDResultXML(resultXML string) ([]AtDMatch, error) {
	// Support both <results><error>...</error></results> and bare fragments.
	type errorNode struct {
		String      string `xml:"string"`
		Description string `xml:"description"`
	}
	type root struct {
		Errors []errorNode `xml:"error"`
	}
	// Try wrapping if needed
	body := strings.TrimSpace(resultXML)
	if !strings.HasPrefix(body, "<") {
		return nil, fmt.Errorf("not XML")
	}
	// decode with flexible root: look for any error elements via token scan
	dec := xml.NewDecoder(strings.NewReader(body))
	var out []AtDMatch
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "error" {
			continue
		}
		var e errorNode
		if err := dec.DecodeElement(&e, &se); err != nil {
			continue
		}
		out = append(out, AtDMatch{Description: e.Description, String: e.String})
	}
	_ = root{} // silence if unused
	return out, nil
}

// AfterTheDeadlineChecker ports org.languagetool.dev.dumpcheck.AfterTheDeadlineChecker
// without live HTTP: Query is injectable.
type AfterTheDeadlineChecker struct {
	URLPrefix        string
	MaxSentenceCount int
	// Query returns AtD XML for a sentence (nil → no matches).
	Query func(text string) (string, error)
}

func NewAfterTheDeadlineChecker(urlPrefix string, maxSentenceCount int) *AfterTheDeadlineChecker {
	return &AfterTheDeadlineChecker{URLPrefix: urlPrefix, MaxSentenceCount: maxSentenceCount}
}

// RunResult is one sentence evaluation line.
type RunResult struct {
	Source  string
	Text    string
	Matches []AtDMatch
}

// Run drains source and queries AtD (or inject). Stops after MaxSentenceCount when >0.
func (c *AfterTheDeadlineChecker) Run(source SentenceSource) ([]RunResult, error) {
	var results []RunResult
	n := 0
	for source.HasNext() {
		sent, err := source.Next()
		if err != nil {
			return results, err
		}
		var matches []AtDMatch
		if c.Query != nil {
			xmlBody, err := c.Query(sent.GetText())
			if err != nil {
				return results, err
			}
			matches, err = ParseAtDResultXML(xmlBody)
			if err != nil {
				return results, err
			}
		}
		results = append(results, RunResult{
			Source:  sent.GetSource(),
			Text:    sent.GetText(),
			Matches: matches,
		})
		n++
		if c.MaxSentenceCount > 0 && n >= c.MaxSentenceCount {
			break
		}
	}
	return results, nil
}
