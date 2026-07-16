package pattern

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/jregex"
)

// LoadFile loads pattern rules from a LanguageTool grammar XML file.
func LoadFile(path string) ([]*Rule, error) {
	r, err := readXMLExpandEntities(path)
	if err != nil {
		return nil, err
	}
	return Load(r)
}

// Load parses grammar XML from r.
func Load(r io.Reader) ([]*Rule, error) {
	dec := xml.NewDecoder(r)
	var rules []*Rule
	var category string
	var categoryDefault string
	var group *ruleGroupState

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rules, fmt.Errorf("xml: %w", err)
		}
		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "category":
				category = attr(se, "id")
				categoryDefault = attr(se, "default")
				if categoryDefault == "" {
					categoryDefault = "on"
				}
			case "rulegroup":
				group = &ruleGroupState{
					ID:      attr(se, "id"),
					Name:    attr(se, "name"),
					Default: attr(se, "default"),
					Cat:     category,
				}
				if group.Default == "" {
					group.Default = categoryDefault
				}
			case "rule":
				rule, err := parseRule(dec, se, category, categoryDefault, group)
				if err != nil {
					return rules, err
				}
				if rule != nil {
					rules = append(rules, rule)
				}
			}
		case xml.EndElement:
			switch se.Name.Local {
			case "rulegroup":
				group = nil
			case "category":
				category = ""
				categoryDefault = "on"
			}
		}
	}
	return rules, nil
}

type ruleGroupState struct {
	ID, Name, Default, Cat string
	Sub                    int
}

func parseRule(dec *xml.Decoder, start xml.StartElement, category, catDefault string, group *ruleGroupState) (*Rule, error) {
	r := &Rule{
		ID:        attr(start, "id"),
		Name:      attr(start, "name"),
		Category:  category,
		Default:   attr(start, "default"),
		IssueType: attr(start, "type"),
	}
	if group != nil {
		if r.ID == "" {
			group.Sub++
			r.ID = group.ID
			r.SubID = strconv.Itoa(group.Sub)
		}
		if r.Name == "" {
			r.Name = group.Name
		}
		if r.Default == "" {
			r.Default = group.Default
		}
		if r.Category == "" {
			r.Category = group.Cat
		}
	}
	if r.Default == "" {
		r.Default = catDefault
	}
	if r.Default == "" {
		r.Default = "on"
	}

	var (
		inPattern    bool
		inAnti       bool
		inMarker     bool
		inMessage    bool
		inSuggestion bool
		inExample    bool
		inShort      bool
		curAnti      []PatToken
		msgBuilder   strings.Builder
		sugBuilder   strings.Builder
		shortBuilder strings.Builder
		exampleRaw   strings.Builder
		exampleCorr  string
		depthPattern int
	)

	// We re-decode the rule element content using a custom walk of tokens until </rule>
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pattern":
				inPattern = true
				depthPattern++
			case "antipattern":
				inAnti = true
				curAnti = nil
			case "marker":
				if inExample {
					exampleRaw.WriteString("<marker>")
				} else {
					inMarker = true
				}
			case "token":
				pt, err := parseToken(dec, t, inMarker)
				if err != nil {
					return nil, err
				}
				if pt.NeedsPOS() {
					r.RequiresPOS = true
				}
				if inAnti {
					curAnti = append(curAnti, pt)
				} else if inPattern {
					r.Tokens = append(r.Tokens, pt)
				}
			case "and", "or":
				// and/or inside <token> is handled in parseToken; at pattern level skip + mark incomplete
				if err := skipElement(dec, t); err != nil {
					return nil, err
				}
				if inPattern {
					r.Incomplete = true
				}
			case "unify", "phraseref", "include":
				if err := skipElement(dec, t); err != nil {
					return nil, err
				}
				r.Incomplete = true
			case "regexp":
				// whole-rule regexp — different matcher
				if err := skipElement(dec, t); err != nil {
					return nil, err
				}
				r.Incomplete = true
			case "filter":
				if err := skipElement(dec, t); err != nil {
					return nil, err
				}
				r.Incomplete = true
			case "message":
				inMessage = true
				msgBuilder.Reset()
			case "short":
				inShort = true
				shortBuilder.Reset()
			case "suggestion":
				if !inMessage {
					inSuggestion = true
					sugBuilder.Reset()
				} else {
					// suggestion inside message — capture as text
					msgBuilder.WriteString("<suggestion>")
				}
			case "example":
				inExample = true
				exampleRaw.Reset()
				exampleCorr = attr(t, "correction")
				// type="correct" etc. — correction empty means correct example when no correction attr...
				// LT uses correction attribute for incorrect examples
			case "url", "tags", "tld", "raw_example":
				if err := skipElement(dec, t); err != nil {
					return nil, err
				}
			default:
				// nested unknown inside message/example: keep text via CharData
				if !inMessage && !inExample && !inSuggestion && !inShort {
					if err := skipElement(dec, t); err != nil {
						return nil, err
					}
				} else if inMessage {
					// serialize match/suggestion tags with attributes for later resolution
					msgBuilder.WriteByte('<')
					msgBuilder.WriteString(t.Name.Local)
					for _, a := range t.Attr {
						msgBuilder.WriteByte(' ')
						msgBuilder.WriteString(a.Name.Local)
						msgBuilder.WriteString(`="`)
						msgBuilder.WriteString(a.Value)
						msgBuilder.WriteByte('"')
					}
					msgBuilder.WriteString("/>")
					// consume empty element body if any
					if err := skipElement(dec, t); err != nil {
						return nil, err
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "pattern":
				depthPattern--
				if depthPattern == 0 {
					inPattern = false
				}
			case "antipattern":
				inAnti = false
				if len(curAnti) > 0 {
					r.Anti = append(r.Anti, curAnti)
				}
			case "marker":
				if inExample {
					exampleRaw.WriteString("</marker>")
				} else {
					inMarker = false
				}
			case "message":
				inMessage = false
				r.Message = strings.TrimSpace(msgBuilder.String())
			case "short":
				inShort = false
				r.ShortMsg = strings.TrimSpace(shortBuilder.String())
			case "suggestion":
				if inSuggestion {
					inSuggestion = false
					r.Suggestions = append(r.Suggestions, strings.TrimSpace(sugBuilder.String()))
				} else if inMessage {
					msgBuilder.WriteString("</suggestion>")
				}
			case "example":
				inExample = false
				ex := parseExample(exampleRaw.String(), exampleCorr)
				r.Examples = append(r.Examples, ex)
			case "rule":
				if r.ID == "" {
					return nil, nil
				}
				// strip suggestion tags from message for display; keep plain text
				r.Message = stripSuggestionTags(r.Message)
				return r, nil
			}
		case xml.CharData:
			s := string(t)
			if inSuggestion {
				sugBuilder.WriteString(s)
			} else if inMessage {
				msgBuilder.WriteString(s)
			} else if inShort {
				shortBuilder.WriteString(s)
			} else if inExample {
				exampleRaw.WriteString(s)
			}
		}
	}
}

func parseToken(dec *xml.Decoder, start xml.StartElement, inMarker bool) (PatToken, error) {
	pt := PatToken{
		CaseSensitive: attr(start, "case_sensitive") == "yes",
		Regexp:        attr(start, "regexp") == "yes",
		Negate:        attr(start, "negate") == "yes",
		Inflected:     attr(start, "inflected") == "yes",
		Postag:        attr(start, "postag"),
		PostagRegexp:  attr(start, "postag_regexp") == "yes",
		Chunk:         attr(start, "chunk"),
		NegatePos:     attr(start, "negate_pos") == "yes",
		SpaceBefore:   attr(start, "spacebefore"),
		InsideMarker:  inMarker,
		Min:           1,
		Max:           1,
	}
	if v := attr(start, "min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pt.Min = n
		}
	}
	if v := attr(start, "max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pt.Max = n
		}
	}
	if v := attr(start, "skip"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pt.Skip = n
		}
	}

	var val strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return pt, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "exception":
				ex, err := parseToken(dec, t, false)
				if err != nil {
					return pt, err
				}
				pt.Exceptions = append(pt.Exceptions, ex)
			case "and":
				// and group: nested tokens
				andToks, err := parseTokenGroup(dec, t)
				if err != nil {
					return pt, err
				}
				pt.And = append(pt.And, andToks...)
			case "or":
				orToks, err := parseTokenGroup(dec, t)
				if err != nil {
					return pt, err
				}
				pt.Or = append(pt.Or, orToks...)
			default:
				if err := skipElement(dec, t); err != nil {
					return pt, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				pt.Value = strings.TrimSpace(val.String())
				if pt.Regexp && pt.Value != "" {
					re, err := jregex.Compile(pt.Value, pt.CaseSensitive)
					if err == nil {
						pt.Re = re
					}
				}
				if cre := attr(start, "chunk_re"); cre != "" {
					if re, err := regexp.Compile("^(?:" + cre + ")$"); err == nil {
						pt.ChunkRe = re
					}
				}
				if pt.PostagRegexp && pt.Postag != "" {
					pat := jregex.JavaToGo(pt.Postag)
					if re, err := regexp.Compile("^(?:" + pat + ")$"); err == nil {
						pt.PostagRe = re
					}
				}
				return pt, nil
			}
		case xml.CharData:
			val.Write(t)
		}
	}
}

func parseTokenGroup(dec *xml.Decoder, start xml.StartElement) ([]PatToken, error) {
	var out []PatToken
	for {
		tok, err := dec.Token()
		if err != nil {
			return out, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "token" {
				pt, err := parseToken(dec, t, false)
				if err != nil {
					return out, err
				}
				out = append(out, pt)
			} else {
				if err := skipElement(dec, t); err != nil {
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

func skipElement(dec *xml.Decoder, start xml.StartElement) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func parseExample(raw, correction string) Example {
	ex := Example{Correction: correction}
	// raw may include <marker>...</marker> (serialized by the loader)
	const open, close = "<marker>", "</marker>"
	if i := strings.Index(raw, open); i >= 0 {
		j := strings.Index(raw, close)
		if j > i {
			inner := raw[i+len(open) : j]
			// MarkerFrom/To are rune offsets in the cleaned example text.
			ex.Text = raw[:i] + inner + raw[j+len(close):]
			ex.HasMarker = true
			ex.MarkerFrom = len([]rune(raw[:i]))
			ex.MarkerTo = ex.MarkerFrom + len([]rune(inner))
			return ex
		}
	}
	ex.Text = raw
	return ex
}

func stripSuggestionTags(s string) string {
	// Preserve match placeholders as $N for later resolution.
	reMatch := regexp.MustCompile(`<match\s+no="(\d+)"[^/]*/>`)
	s = reMatch.ReplaceAllString(s, `$$$1`)
	s = strings.ReplaceAll(s, "<suggestion>", "'")
	s = strings.ReplaceAll(s, "</suggestion>", "'")
	re := regexp.MustCompile(`</?[^>]+>`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}
