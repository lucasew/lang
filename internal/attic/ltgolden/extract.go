package ltgolden

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/attic/pattern"
)

// xmlExample maps LT <example> using encoding/xml.
type xmlExample struct {
	Correction string `xml:"correction,attr"`
	Type       string `xml:"type,attr"` // correct | ambiguous | untouched | …
	InnerXML   string `xml:",innerxml"`
}

type xmlRule struct {
	ID       string       `xml:"id,attr"`
	Default  string       `xml:"default,attr"`
	Examples []xmlExample `xml:"example"`
}

// ExtractCases loads ALL LT grammar/style XML examples (no rule filtering).
func ExtractCases(grammarPaths []string) ([]Case, error) {
	var out []Case
	for _, p := range grammarPaths {
		cases, err := extractGrammarFile(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, cases...)
	}
	return out, nil
}

func extractGrammarFile(path string) ([]Case, error) {
	r, err := pattern.OpenExpandedXML(path)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(r)
	lang := langFromRulesPath(path)

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
			cases, err := readRuleGroup(dec, se, path, lang)
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
				id = "ANON"
			}
			out = append(out, casesFromGrammarExamples(id, rule.Default, rule.Examples, path, lang)...)
		}
	}
	return out, nil
}

func readRuleGroup(dec *xml.Decoder, start xml.StartElement, path, lang string) ([]Case, error) {
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
				out = append(out, casesFromGrammarExamples(id, def, rule.Examples, path, lang)...)
			case "example":
				var ex xmlExample
				if err := dec.DecodeElement(&ex, &t); err != nil {
					return out, err
				}
				if groupID != "" {
					out = append(out, casesFromGrammarExamples(groupID, groupDefault, []xmlExample{ex}, path, lang)...)
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

func casesFromGrammarExamples(ruleID, def string, examples []xmlExample, path, lang string) []Case {
	var out []Case
	for _, ex := range examples {
		parsed := parseMarkedInner(ex.InnerXML)
		text := strings.TrimSpace(parsed.text)
		if text == "" {
			continue
		}
		// LT: correction attr => incorrect example; absence => correct (should not match)
		// type="correct" also used rarely; type empty + no correction = correct
		incorrect := ex.Correction != ""
		if strings.EqualFold(ex.Type, "correct") {
			incorrect = false
		}
		out = append(out, Case{
			Kind:        KindGrammarExample,
			Lang:        lang,
			RuleID:      ruleID,
			RuleDefault: def,
			Text:        text,
			Incorrect:   incorrect,
			Correction:  ex.Correction,
			HasMarker:   parsed.hasMarker,
			MarkerFrom:  parsed.from,
			MarkerTo:    parsed.to,
			SourceFile:  path,
			ExampleType: ex.Type,
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
			cleaned := stripTags(raw[:i] + inner + raw[j+len(close):])
			prefix := stripTags(raw[:i])
			innerClean := stripTags(inner)
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

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func langFromRulesPath(path string) string {
	// .../languagetool-language-modules/<lang>/.../rules/<lang>/file.xml
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, p := range parts {
		if p == "languagetool-language-modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "und"
}

var _ = pattern.OpenExpandedXML
