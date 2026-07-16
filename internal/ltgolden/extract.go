package ltgolden

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/pattern"
)

// xmlExample maps LT <example correction="…">…<marker>…</marker>…</example>
// using encoding/xml struct tags.
type xmlExample struct {
	Correction string `xml:"correction,attr"`
	InnerXML   string `xml:",innerxml"`
}

// xmlRule is the subset of <rule> we need for goldens.
type xmlRule struct {
	ID       string       `xml:"id,attr"`
	Default  string       `xml:"default,attr"`
	Examples []xmlExample `xml:"example"`
}

// ExtractCases loads LT grammar XML examples via encoding/xml.Decoder + DecodeElement.
func ExtractCases(grammarPaths []string) ([]Case, error) {
	var out []Case
	for _, p := range grammarPaths {
		cases, err := extractFile(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, cases...)
	}
	return out, nil
}

func extractFile(path string) ([]Case, error) {
	// Expand DTD entities then parse with the standard library decoder.
	r, err := pattern.OpenExpandedXML(path)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(r)

	var out []Case
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
			cases, err := readRuleGroup(dec, se, path)
			if err != nil {
				return out, err
			}
			out = append(out, cases...)
		case "rule":
			var rule xmlRule
			if err := dec.DecodeElement(&rule, &se); err != nil {
				return out, err
			}
			id := rule.ID
			if id == "" {
				continue
			}
			def := rule.Default
			if def == "" {
				def = "on"
			}
			if skipRule(id, def) {
				continue
			}
			out = append(out, casesFromExamples(id, rule.Examples, path)...)
		}
	}
	return out, nil
}

func readRuleGroup(dec *xml.Decoder, start xml.StartElement, path string) ([]Case, error) {
	groupID := attr(start, "id")
	groupDefault := attr(start, "default")
	groupSub := 0
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
				var rule xmlRule
				if err := dec.DecodeElement(&rule, &t); err != nil {
					return out, err
				}
				id := rule.ID
				if id == "" {
					groupSub++
					id = fmt.Sprintf("%s[%d]", groupID, groupSub)
				}
				def := rule.Default
				if def == "" {
					def = groupDefault
				}
				if def == "" {
					def = "on"
				}
				if skipRule(id, def) {
					continue
				}
				out = append(out, casesFromExamples(id, rule.Examples, path)...)
			case "example":
				var ex xmlExample
				if err := dec.DecodeElement(&ex, &t); err != nil {
					return out, err
				}
				if groupID != "" && !skipRule(groupID, orOn(groupDefault)) {
					out = append(out, casesFromExamples(groupID, []xmlExample{ex}, path)...)
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

func orOn(d string) string {
	if d == "" {
		return "on"
	}
	return d
}

func casesFromExamples(ruleID string, examples []xmlExample, path string) []Case {
	var out []Case
	for _, ex := range examples {
		parsed := parseMarkedInner(ex.InnerXML)
		text := strings.TrimSpace(parsed.text)
		if text == "" {
			continue
		}
		out = append(out, Case{
			RuleID:     ruleID,
			Text:       text,
			Incorrect:  ex.Correction != "",
			Correction: ex.Correction,
			HasMarker:  parsed.hasMarker,
			MarkerFrom: parsed.from,
			MarkerTo:   parsed.to,
			SourceFile: path,
		})
	}
	return out
}

type marked struct {
	text      string
	hasMarker bool
	from, to  int
}

func parseMarkedInner(raw string) marked {
	const open, close = "<marker>", "</marker>"
	if i := strings.Index(raw, open); i >= 0 {
		j := strings.Index(raw, close)
		if j > i {
			inner := raw[i+len(open) : j]
			cleaned := raw[:i] + inner + raw[j+len(close):]
			// strip any other tags that may appear rarely
			cleaned = stripTags(cleaned)
			innerClean := stripTags(inner)
			prefix := stripTags(raw[:i])
			return marked{
				text:      cleaned,
				hasMarker: true,
				from:      len([]rune(prefix)),
				to:        len([]rune(prefix)) + len([]rune(innerClean)),
			}
		}
	}
	return marked{text: stripTags(raw)}
}

func stripTags(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '<':
			in = true
		case '>':
			in = false
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func skipRule(id, def string) bool {
	if def == "off" || def == "temp_off" {
		return true
	}
	return strings.HasPrefix(id, "AI_") || strings.HasPrefix(id, "QB_")
}

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// silence unused in case of build tags
var _ = filepath.Separator
var _ = pattern.OpenExpandedXML
