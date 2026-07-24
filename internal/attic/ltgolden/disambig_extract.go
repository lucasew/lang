package ltgolden

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/attic/pattern"
)

// Disambiguation examples use type=ambiguous|untouched and inputform/outputform attrs.
type xmlDisambigExample struct {
	Type       string `xml:"type,attr"`
	InputForm  string `xml:"inputform,attr"`
	OutputForm string `xml:"outputform,attr"`
	InnerXML   string `xml:",innerxml"`
}

type xmlDisambigRule struct {
	ID       string               `xml:"id,attr"`
	Examples []xmlDisambigExample `xml:"example"`
}

// ExtractDisambigCases loads ALL disambiguation.xml examples (no skips).
func ExtractDisambigCases(paths []string) ([]Case, error) {
	var out []Case
	for _, p := range paths {
		cs, err := extractDisambigFile(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, cs...)
	}
	return out, nil
}

func extractDisambigFile(path string) ([]Case, error) {
	r, err := pattern.OpenExpandedXML(path)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(r)
	lang := langFromResourcePath(path)
	var out []Case
	groupID := ""
	groupSub := 0

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "rulegroup":
			groupID = attr(se, "id")
			groupSub = 0
			cs, err := readDisambigGroup(dec, se, path, lang, groupID, &groupSub)
			if err != nil {
				return out, err
			}
			out = append(out, cs...)
			groupID = ""
		case "rule":
			var rule xmlDisambigRule
			if err := dec.DecodeElement(&rule, &se); err != nil {
				return out, err
			}
			id := rule.ID
			if id == "" {
				id = "ANON"
			}
			out = append(out, casesFromDisambig(id, rule.Examples, path, lang)...)
		}
	}
	return out, nil
}

func readDisambigGroup(dec *xml.Decoder, start xml.StartElement, path, lang, groupID string, groupSub *int) ([]Case, error) {
	var out []Case
	for {
		tok, err := dec.Token()
		if err != nil {
			return out, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rule":
				var rule xmlDisambigRule
				if err := dec.DecodeElement(&rule, &t); err != nil {
					return out, err
				}
				id := rule.ID
				if id == "" {
					*groupSub++
					id = fmt.Sprintf("%s[%d]", groupID, *groupSub)
				}
				out = append(out, casesFromDisambig(id, rule.Examples, path, lang)...)
			case "example":
				var ex xmlDisambigExample
				if err := dec.DecodeElement(&ex, &t); err != nil {
					return out, err
				}
				if groupID != "" {
					out = append(out, casesFromDisambig(groupID, []xmlDisambigExample{ex}, path, lang)...)
				}
			default:
				if err := dec.Skip(); err != nil {
					return out, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return out, nil
			}
		}
	}
}

func casesFromDisambig(ruleID string, examples []xmlDisambigExample, path, lang string) []Case {
	var out []Case
	for _, ex := range examples {
		parsed := parseMarkedInner(ex.InnerXML)
		text := strings.TrimSpace(parsed.text)
		if text == "" {
			continue
		}
		// untouched => analysis must not change for marked token (correct-like)
		// ambiguous => expect outputform readings after disambig
		incorrect := !strings.EqualFold(ex.Type, "untouched")
		out = append(out, Case{
			Kind:        KindDisambigExample,
			Lang:        lang,
			RuleID:      ruleID,
			Text:        text,
			Incorrect:   incorrect,
			Correction:  ex.OutputForm, // reuse field for expected outputform
			HasMarker:   parsed.hasMarker,
			MarkerFrom:  parsed.from,
			MarkerTo:    parsed.to,
			SourceFile:  path,
			ExampleType: ex.Type + "|" + ex.InputForm + "=>" + ex.OutputForm,
		})
	}
	return out
}

func langFromResourcePath(path string) string {
	// .../resource/<lang>/disambiguation.xml
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, p := range parts {
		if p == "resource" && i+1 < len(parts) {
			return parts[i+1]
		}
		if p == "languagetool-language-modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "und"
}
